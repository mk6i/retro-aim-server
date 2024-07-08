package state

import (
	"context"
	"log/slog"
	"sync"

	"github.com/mk6i/retro-aim-server/wire"
)

// InMemorySessionManager handles the lifecycle of a user session and provides
// synchronized message relay between sessions in the session pool. An
// InMemorySessionManager is safe for concurrent use by multiple goroutines.
type InMemorySessionManager struct {
	store    map[IdentScreenName]*Session
	mapMutex sync.RWMutex
	logger   *slog.Logger
}

// NewInMemorySessionManager creates a new instance of InMemorySessionManager.
func NewInMemorySessionManager(logger *slog.Logger) *InMemorySessionManager {
	return &InMemorySessionManager{
		logger: logger,
		store:  make(map[IdentScreenName]*Session),
	}
}

// RelayToAll relays a message to all sessions in the session pool.
func (s *InMemorySessionManager) RelayToAll(ctx context.Context, msg wire.SNACMessage) {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	for _, sess := range s.store {
		s.maybeRelayMessage(ctx, msg, sess)
	}
}

// RelayToScreenName relays a message to a session with a matching screen name.
func (s *InMemorySessionManager) RelayToScreenName(ctx context.Context, screenName IdentScreenName, msg wire.SNACMessage) {
	sess := s.RetrieveByScreenName(screenName)
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

// AddSession adds a new session to the pool. It replaces an existing session
// with a matching screen name, ensuring that each screen name is unique in the
// pool. This method does not return a nil session value. It's possible to add
// a non-existent user to the session. Callers should ensure that the account
// represented by displayScreenName is valid.
func (s *InMemorySessionManager) AddSession(displayScreenName DisplayScreenName) *Session {
	s.mapMutex.Lock()
	defer s.mapMutex.Unlock()

	identScreenName := NewIdentScreenName(string(displayScreenName))
	// Only allow one session at a time per screen name. A session may already
	// exist because:
	// 1) the user is signing on using an already logged-on screen name.
	// 2) the session might be orphaned due to an undetected client
	// disconnection.
	for _, sess := range s.store {
		if identScreenName == sess.IdentScreenName() {
			sess.Close()
			delete(s.store, identScreenName)
			break
		}
	}

	sess := NewSession()
	sess.SetIdentScreenName(identScreenName)
	sess.SetDisplayScreenName(displayScreenName)
	s.store[identScreenName] = sess
	return sess
}

// RemoveSession takes a session out of the session pool.
func (s *InMemorySessionManager) RemoveSession(sess *Session) {
	s.mapMutex.Lock()
	defer s.mapMutex.Unlock()
	delete(s.store, sess.IdentScreenName())
}

// RetrieveSession finds a session with a matching sessionID. Returns nil if
// session is not found.
func (s *InMemorySessionManager) RetrieveSession(screenName IdentScreenName) *Session {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	return s.store[screenName]
}

// RetrieveByScreenName find a session with a matching screen name. Returns nil
// if session is not found.
func (s *InMemorySessionManager) RetrieveByScreenName(screenName IdentScreenName) *Session {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	for _, sess := range s.store {
		if screenName == sess.IdentScreenName() {
			return sess
		}
	}
	return nil
}

func (s *InMemorySessionManager) retrieveByScreenNames(screenNames []IdentScreenName) []*Session {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	var ret []*Session
	for _, sn := range screenNames {
		for _, sess := range s.store {
			if sn == sess.IdentScreenName() {
				ret = append(ret, sess)
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
	for _, sess := range s.store {
		sessions = append(sessions, sess)
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

// AddSession adds a user to a chat room.
func (s *InMemoryChatSessionManager) AddSession(chatCookie string, screenName DisplayScreenName) *Session {
	s.mapMutex.Lock()
	defer s.mapMutex.Unlock()

	if _, ok := s.store[chatCookie]; !ok {
		s.store[chatCookie] = NewInMemorySessionManager(s.logger)
	}

	sessionManager := s.store[chatCookie]
	sess := sessionManager.AddSession(screenName)
	sess.SetChatRoomCookie(chatCookie)
	return sess
}

// RemoveSession removes a user session from a chat room.
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
