package state

import (
	"context"
	"log/slog"
	"sync"

	"github.com/mkaminski/goaim/oscar"
)

type InMemorySessionManager struct {
	store    map[string]*Session
	mapMutex sync.RWMutex
	logger   *slog.Logger
}

func NewSessionManager(logger *slog.Logger) *InMemorySessionManager {
	return &InMemorySessionManager{
		logger: logger,
		store:  make(map[string]*Session),
	}
}

func (s *InMemorySessionManager) Broadcast(ctx context.Context, msg oscar.SNACMessage) {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	for _, sess := range s.store {
		s.maybeSendMessage(ctx, msg, sess)
	}
}

func (s *InMemorySessionManager) maybeSendMessage(ctx context.Context, msg oscar.SNACMessage, sess *Session) {
	switch sess.SendMessage(msg) {
	case SessSendClosed:
		s.logger.WarnContext(ctx, "can't send notification because the user's session is closed", "recipient", sess.ScreenName(), "message", msg)
	case SessQueueFull:
		s.logger.WarnContext(ctx, "can't send notification because queue is full", "recipient", sess.ScreenName(), "message", msg)
		sess.Close()
	}
}

func (s *InMemorySessionManager) Empty() bool {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	return len(s.store) == 0
}

func (s *InMemorySessionManager) Participants() []*Session {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	var sessions []*Session
	for _, sess := range s.store {
		sessions = append(sessions, sess)
	}
	return sessions
}

func (s *InMemorySessionManager) BroadcastExcept(ctx context.Context, except *Session, msg oscar.SNACMessage) {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	for _, sess := range s.store {
		if sess == except {
			continue
		}
		s.maybeSendMessage(ctx, msg, sess)
	}
}

func (s *InMemorySessionManager) Retrieve(ID string) *Session {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	return s.store[ID]
}

func (s *InMemorySessionManager) RetrieveByScreenName(screenName string) *Session {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	for _, sess := range s.store {
		if screenName == sess.ScreenName() {
			return sess
		}
	}
	return nil
}

func (s *InMemorySessionManager) retrieveByScreenNames(screenNames []string) []*Session {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	var ret []*Session
	for _, sn := range screenNames {
		for _, sess := range s.store {
			if sn == sess.ScreenName() {
				ret = append(ret, sess)
			}
		}
	}
	return ret
}

func (s *InMemorySessionManager) SendToScreenName(ctx context.Context, screenName string, msg oscar.SNACMessage) {
	sess := s.RetrieveByScreenName(screenName)
	if sess == nil {
		s.logger.WarnContext(ctx, "can't send notification because user is not online", "recipient", screenName, "message", msg)
		return
	}
	s.maybeSendMessage(ctx, msg, sess)
}

func (s *InMemorySessionManager) BroadcastToScreenNames(ctx context.Context, screenNames []string, msg oscar.SNACMessage) {
	for _, sess := range s.retrieveByScreenNames(screenNames) {
		s.maybeSendMessage(ctx, msg, sess)
	}
}

func (s *InMemorySessionManager) NewSessionWithSN(sessID string, screenName string) *Session {
	s.mapMutex.Lock()
	defer s.mapMutex.Unlock()

	// Only allow one session at a time per screen name. A session may already
	// exist because:
	// 1) the user is signing on using an already logged-on screen name.
	// 2) the session might be orphaned due to an undetected client
	// disconnection.
	for _, sess := range s.store {
		if screenName == sess.ScreenName() {
			sess.Close()
			delete(s.store, sess.ID())
			break
		}
	}

	sess := NewSession()
	sess.SetID(sessID)
	sess.SetScreenName(screenName)
	s.store[sess.ID()] = sess
	return sess
}

func (s *InMemorySessionManager) Remove(sess *Session) {
	s.mapMutex.Lock()
	defer s.mapMutex.Unlock()
	delete(s.store, sess.ID())
}
