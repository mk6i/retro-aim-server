package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/mkaminski/goaim/user"
	"log/slog"
	"sync"
	"time"

	"github.com/mkaminski/goaim/oscar"
)

var (
	ErrSessNotFound = errors.New("session was not found")
	ErrSignedOff    = errors.New("user signed off")
)

type InMemorySessionManager struct {
	store    map[string]*user.Session
	mapMutex sync.RWMutex
	logger   *slog.Logger
}

func NewSessionManager(logger *slog.Logger) *InMemorySessionManager {
	return &InMemorySessionManager{
		logger: logger,
		store:  make(map[string]*user.Session),
	}
}

func (s *InMemorySessionManager) Broadcast(ctx context.Context, msg oscar.XMessage) {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	for _, sess := range s.store {
		s.maybeSendMessage(ctx, msg, sess)
	}
}

func (s *InMemorySessionManager) maybeSendMessage(ctx context.Context, msg oscar.XMessage, sess *user.Session) {
	switch sess.SendMessage(msg) {
	case user.SessSendClosed:
		s.logger.WarnContext(ctx, "can't send notification because the user's session is closed", "recipient", sess.ScreenName(), "message", msg)
	case user.SessQueueFull:
		s.logger.WarnContext(ctx, "can't send notification because queue is full", "recipient", sess.ScreenName(), "message", msg)
		sess.Close()
	}
}

func (s *InMemorySessionManager) Empty() bool {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	return len(s.store) == 0
}

func (s *InMemorySessionManager) Participants() []*user.Session {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	var sessions []*user.Session
	for _, sess := range s.store {
		sessions = append(sessions, sess)
	}
	return sessions
}

func (s *InMemorySessionManager) BroadcastExcept(ctx context.Context, except *user.Session, msg oscar.XMessage) {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	for _, sess := range s.store {
		if sess == except {
			continue
		}
		s.maybeSendMessage(ctx, msg, sess)
	}
}

func (s *InMemorySessionManager) Retrieve(ID string) (*user.Session, bool) {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	sess, found := s.store[ID]
	return sess, found
}

func (s *InMemorySessionManager) RetrieveByScreenName(screenName string) (*user.Session, error) {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	for _, sess := range s.store {
		if screenName == sess.ScreenName() {
			return sess, nil
		}
	}
	return nil, fmt.Errorf("%w: %s", ErrSessNotFound, screenName)
}

func (s *InMemorySessionManager) retrieveByScreenNames(screenNames []string) []*user.Session {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	var ret []*user.Session
	for _, sn := range screenNames {
		for _, sess := range s.store {
			if sn == sess.ScreenName() {
				ret = append(ret, sess)
			}
		}
	}
	return ret
}

func (s *InMemorySessionManager) SendToScreenName(ctx context.Context, screenName string, msg oscar.XMessage) {
	sess, err := s.RetrieveByScreenName(screenName)
	if err != nil {
		s.logger.WarnContext(ctx, "can't send notification because user is not online", "recipient", screenName, "message", msg)
		return
	}
	s.maybeSendMessage(ctx, msg, sess)
}

func (s *InMemorySessionManager) BroadcastToScreenNames(ctx context.Context, screenNames []string, msg oscar.XMessage) {
	for _, sess := range s.retrieveByScreenNames(screenNames) {
		s.maybeSendMessage(ctx, msg, sess)
	}
}

func (s *InMemorySessionManager) NewSessionWithSN(sessID string, screenName string) *user.Session {
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

	sess := user.NewSession()
	sess.SetID(sessID)
	sess.SetScreenName(screenName)
	s.store[sess.ID()] = sess
	return sess
}

func (s *InMemorySessionManager) Remove(sess *user.Session) {
	s.mapMutex.Lock()
	defer s.mapMutex.Unlock()
	delete(s.store, sess.ID())
}

type ChatRoom struct {
	CreateTime     time.Time
	DetailLevel    uint8
	Exchange       uint16
	Cookie         string
	InstanceNumber uint16
	Name           string
	SessionManager
}

func (c ChatRoom) TLVList() []oscar.TLV {
	return []oscar.TLV{
		oscar.NewTLV(0x00c9, uint16(15)),
		oscar.NewTLV(0x00ca, uint32(c.CreateTime.Unix())),
		oscar.NewTLV(0x00d1, uint16(1024)),
		oscar.NewTLV(0x00d2, uint16(100)),
		oscar.NewTLV(0x00d5, uint8(2)),
		oscar.NewTLV(0x006a, c.Name),
		oscar.NewTLV(0x00d3, c.Name),
	}
}

type ChatRegistry struct {
	store    map[string]ChatRoom
	mapMutex sync.RWMutex
}

func NewChatRegistry() *ChatRegistry {
	return &ChatRegistry{
		store: make(map[string]ChatRoom),
	}
}

func (c *ChatRegistry) Register(room ChatRoom) {
	c.mapMutex.Lock()
	defer c.mapMutex.Unlock()
	c.store[room.Cookie] = room
}

func (c *ChatRegistry) Retrieve(chatID string) (ChatRoom, error) {
	c.mapMutex.RLock()
	defer c.mapMutex.RUnlock()
	sm, found := c.store[chatID]
	if !found {
		return sm, errors.New("unable to find session manager for chat")
	}
	return sm, nil
}

func (c *ChatRegistry) MaybeRemoveRoom(chatID string) {
	c.mapMutex.Lock()
	defer c.mapMutex.Unlock()

	room, found := c.store[chatID]
	if found && room.Empty() {
		delete(c.store, chatID)
	}
}
