package state

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/mk6i/retro-aim-server/wire"
)

type sessionSlot struct {
	sess    *Session
	removed chan bool
}

var errSessConflict = errors.New("session conflict: another session was created concurrently for this user")

// InMemorySessionManager handles the lifecycle of a user session and provides
// synchronized message relay between sessions in the session pool. An
// InMemorySessionManager is safe for concurrent use by multiple goroutines.
type InMemorySessionManager struct {
	store    map[IdentScreenName]*sessionSlot
	mapMutex sync.RWMutex
	logger   *slog.Logger
}

// NewInMemorySessionManager creates a new instance of InMemorySessionManager.
func NewInMemorySessionManager(logger *slog.Logger) *InMemorySessionManager {
	return &InMemorySessionManager{
		logger: logger,
		store:  make(map[IdentScreenName]*sessionSlot),
	}
}

// RelayToAll relays a message to all sessions in the session pool.
func (s *InMemorySessionManager) RelayToAll(ctx context.Context, msg wire.SNACMessage) {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	for _, rec := range s.store {
		s.maybeRelayMessage(ctx, msg, rec.sess)
	}
}

// RelayToScreenName relays a message to a session with a matching screen name.
func (s *InMemorySessionManager) RelayToScreenName(ctx context.Context, screenName IdentScreenName, msg wire.SNACMessage) {
	sess := s.RetrieveSession(screenName)
	if sess == nil {
		s.logger.WarnContext(ctx, "can't send notification because user is not online", "recipient", screenName, "message", msg)
		return
	}
	s.maybeRelayMessage(ctx, msg, sess)
}

// RelayToScreenNames relays a message to sessions with matching screenNames.
func (s *InMemorySessionManager) RelayToScreenNames(ctx context.Context, screenNames []IdentScreenName, msg wire.SNACMessage) {
	for _, sess := range s.retrieveByScreenNames(screenNames) {
		s.maybeRelayMessage(ctx, msg, sess)
	}
}

func (s *InMemorySessionManager) maybeRelayMessage(ctx context.Context, msg wire.SNACMessage, sess *Session) {
	switch sess.RelayMessage(msg) {
	case SessSendClosed:
		s.logger.WarnContext(ctx, "can't send notification because the user's session is closed", "recipient", sess.IdentScreenName(), "message", msg)
	case SessQueueFull:
		s.logger.WarnContext(ctx, "can't send notification because queue is full", "recipient", sess.IdentScreenName(), "message", msg)
		sess.Close()
	}
}

// AddSession adds a new session to the pool, ensuring only one session exists
// for a given screen name. If a session with the same screen name is already
// active, the call blocks until the active session is terminated by
// [InMemorySessionManager.RemoveSession] or the context is canceled. When
// concurrent calls are made for the same screen name, only one call succeeds
// and the others return an error.
func (s *InMemorySessionManager) AddSession(ctx context.Context, screenName DisplayScreenName) (*Session, error) {
	s.mapMutex.Lock()

	active := s.findRec(screenName.IdentScreenName())
	if active != nil {
		// there's an active session that needs to be removed. don't hold the
		// lock while we wait.
		s.mapMutex.Unlock()

		// signal to callers that this session has to go
		active.sess.Close()

		select {
		case <-active.removed: // wait for RemoveSession to be called
		case <-ctx.Done():
			return nil, fmt.Errorf("waiting for previous session to terminate: %w", ctx.Err())
		}

		// the session has been removed, let's try to replace it
		s.mapMutex.Lock()
	}

	defer s.mapMutex.Unlock()

	// make sure a concurrent call didn't already add a session
	if active != nil && s.findRec(screenName.IdentScreenName()) != nil {
		return nil, errSessConflict
	}

	sess := NewSession()
	sess.SetIdentScreenName(screenName.IdentScreenName())
	sess.SetDisplayScreenName(screenName)

	s.store[sess.IdentScreenName()] = &sessionSlot{
		sess:    sess,
		removed: make(chan bool),
	}

	return sess, nil
}

func (s *InMemorySessionManager) findRec(identScreenName IdentScreenName) *sessionSlot {
	for _, rec := range s.store {
		if identScreenName == rec.sess.IdentScreenName() {
			return rec
		}
	}
	return nil
}

// RemoveSession takes a session out of the session pool.
func (s *InMemorySessionManager) RemoveSession(sess *Session) {
	s.mapMutex.Lock()
	defer s.mapMutex.Unlock()
	if rec, ok := s.store[sess.IdentScreenName()]; ok && rec.sess == sess {
		delete(s.store, sess.IdentScreenName())
		close(rec.removed)
	}
}

// RetrieveSession finds a session with a matching sessionID. Returns nil if
// session is not found.
func (s *InMemorySessionManager) RetrieveSession(screenName IdentScreenName) *Session {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	if rec, ok := s.store[screenName]; ok {
		return rec.sess
	}
	return nil
}

func (s *InMemorySessionManager) retrieveByScreenNames(screenNames []IdentScreenName) []*Session {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	var ret []*Session
	for _, sn := range screenNames {
		for _, rec := range s.store {
			if sn == rec.sess.IdentScreenName() {
				ret = append(ret, rec.sess)
			}
		}
	}
	return ret
}

// Empty returns true if the session pool contains 0 sessions.
func (s *InMemorySessionManager) Empty() bool {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	return len(s.store) == 0
}

// AllSessions returns all sessions in the session pool.
func (s *InMemorySessionManager) AllSessions() []*Session {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	var sessions []*Session
	for _, rec := range s.store {
		sessions = append(sessions, rec.sess)
	}
	return sessions
}

// NewInMemoryChatSessionManager creates a new instance of
// InMemoryChatSessionManager.
func NewInMemoryChatSessionManager(logger *slog.Logger) *InMemoryChatSessionManager {
	return &InMemoryChatSessionManager{
		store:  make(map[string]*InMemorySessionManager),
		logger: logger,
	}
}

// InMemoryChatSessionManager manages chat sessions for multiple chat rooms
// stored in memory. It provides thread-safe operations to add, remove, and
// manipulate sessions as well as relay messages to participants.
type InMemoryChatSessionManager struct {
	logger   *slog.Logger
	mapMutex sync.RWMutex
	store    map[string]*InMemorySessionManager
}

// AddSession adds a user to a chat room. If screenName already exists, the old
// session is replaced by a new one.
func (s *InMemoryChatSessionManager) AddSession(ctx context.Context, chatCookie string, screenName DisplayScreenName) (*Session, error) {
	s.mapMutex.Lock()
	defer s.mapMutex.Unlock()

	if _, ok := s.store[chatCookie]; !ok {
		s.store[chatCookie] = NewInMemorySessionManager(s.logger)
	}

	sessionManager := s.store[chatCookie]

	sess, err := sessionManager.AddSession(ctx, screenName)
	if err != nil {
		return nil, fmt.Errorf("AddSession: %w", err)
	}

	sess.SetChatRoomCookie(chatCookie)

	return sess, nil
}

// RemoveSession removes a user session from a chat room. It panics if you
// attempt to remove the session twice.
func (s *InMemoryChatSessionManager) RemoveSession(sess *Session) {
	s.mapMutex.Lock()
	defer s.mapMutex.Unlock()

	sessionManager, ok := s.store[sess.ChatRoomCookie()]
	if !ok {
		panic("attempting to remove a session after its room has been deleted")
	}
	sessionManager.RemoveSession(sess)

	if sessionManager.Empty() {
		delete(s.store, sess.ChatRoomCookie())
	}
}

// AllSessions returns all chat room participants. Returns
// ErrChatRoomNotFound if the room does not exist.
func (s *InMemoryChatSessionManager) AllSessions(cookie string) []*Session {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()

	sessionManager, ok := s.store[cookie]
	if !ok {
		s.logger.Debug("trying to get sessions for non-existent room", "cookie", cookie)
		return nil
	}
	return sessionManager.AllSessions()
}

// RelayToAllExcept sends a message to all chat room participants except for
// the participant with a particular screen name. Returns ErrChatRoomNotFound
// if the room does not exist for cookie.
func (s *InMemoryChatSessionManager) RelayToAllExcept(ctx context.Context, cookie string, except IdentScreenName, msg wire.SNACMessage) {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()

	sessionManager, ok := s.store[cookie]
	if !ok {
		s.logger.Error("trying to relay message to all for non-existent room", "cookie", cookie)
		return
	}

	for _, sess := range sessionManager.AllSessions() {
		if sess.IdentScreenName() == except {
			continue
		}
		sessionManager.maybeRelayMessage(ctx, msg, sess)
	}
}

// RelayToScreenName sends a message to a chat room user. Returns
// ErrChatRoomNotFound if the room does not exist for cookie.
func (s *InMemoryChatSessionManager) RelayToScreenName(ctx context.Context, cookie string, recipient IdentScreenName, msg wire.SNACMessage) {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()

	sessionManager, ok := s.store[cookie]
	if !ok {
		s.logger.Error("trying to relay message to screen name for non-existent room", "cookie", cookie)
		return
	}
	sessionManager.RelayToScreenName(ctx, recipient, msg)
}
