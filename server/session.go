package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/mkaminski/goaim/oscar"
	"log/slog"
	"sync"
	"time"
)

var (
	ErrSessNotFound = errors.New("session was not found")
	ErrSignedOff    = errors.New("user signed off")
)

type SessSendStatus int

const (
	// SessSendOK indicates message was sent to recipient
	SessSendOK SessSendStatus = iota
	// SessSendClosed indicates send did not complete because session is closed
	SessSendClosed
	// SessSendTimeout indicates send timed out due to blocked recipient
	SessSendTimeout
)

const sendTimeout = 10 * time.Second

type Session struct {
	ID          string
	ScreenName  string
	msgCh       chan XMessage
	stopCh      chan struct{}
	Mutex       sync.RWMutex
	Warning     uint16
	AwayMessage string
	SignonTime  time.Time
	invisible   bool
	idle        bool
	idleTime    time.Time
	sendTimeout time.Duration
	closed      bool
}

func (s *Session) IncreaseWarning(incr uint16) {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	s.Warning += incr
}

func (s *Session) SetInvisible(invisible bool) {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	s.invisible = invisible
}

func (s *Session) Invisible() bool {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return s.invisible
}

func (s *Session) SetIdle(dur time.Duration) {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	s.idle = true
	// set the time the user became idle
	s.idleTime = time.Now().Add(-dur)
}

func (s *Session) SetActive() {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	s.idle = false
}

func (s *Session) Idle() bool {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return s.idle
}

func (s *Session) SetAwayMessage(awayMessage string) {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	s.AwayMessage = awayMessage
}

func (s *Session) GetAwayMessage() string {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	return s.AwayMessage
}

func (s *Session) GetTLVUserInfo() oscar.TLVUserInfo {
	return oscar.TLVUserInfo{
		ScreenName:   s.ScreenName,
		WarningLevel: s.GetWarning(),
		TLVBlock: oscar.TLVBlock{
			TLVList: s.GetUserInfo(),
		},
	}
}

func (s *Session) GetUserInfo() oscar.TLVList {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	// sign-in timestamp
	tlvs := oscar.TLVList{}

	tlvs.AddTLV(oscar.NewTLV(0x03, uint32(s.SignonTime.Unix())))

	// away message status
	if s.AwayMessage != "" {
		tlvs.AddTLV(oscar.NewTLV(0x01, uint16(0x0010)|uint16(0x0020)))
	} else {
		tlvs.AddTLV(oscar.NewTLV(0x01, uint16(0x0010)))
	}

	// invisibility status
	if s.invisible {
		tlvs.AddTLV(oscar.NewTLV(0x06, uint16(0x0100)))
	} else {
		tlvs.AddTLV(oscar.NewTLV(0x06, uint16(0x0000)))
	}

	// idle status
	if s.idle {
		tlvs.AddTLV(oscar.NewTLV(0x04, uint16(time.Now().Sub(s.idleTime).Seconds())))
	} else {
		tlvs.AddTLV(oscar.NewTLV(0x04, uint16(0)))
	}

	// capabilities
	var caps []byte
	// chat capability
	caps = append(caps, CapChat...)
	tlvs.AddTLV(oscar.NewTLV(0x0D, caps))

	return tlvs
}

func (s *Session) GetWarning() uint16 {
	var w uint16
	s.Mutex.RLock()
	w = s.Warning
	s.Mutex.RUnlock()
	return w
}

func (s *Session) RecvMessage() chan XMessage {
	return s.msgCh
}

func (s *Session) SendMessage(msg XMessage) SessSendStatus {
	select {
	case s.msgCh <- msg:
		return SessSendOK
	case <-s.stopCh:
		return SessSendClosed
	case <-time.After(s.sendTimeout):
		return SessSendTimeout
	}
}

func (s *Session) Close() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	if s.closed {
		return
	}
	close(s.stopCh)
	s.closed = true
}

func (s *Session) Closed() <-chan struct{} {
	return s.stopCh
}

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

func (s *InMemorySessionManager) Broadcast(ctx context.Context, msg XMessage) {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	for _, sess := range s.store {
		go s.maybeSendMessage(ctx, msg, sess)
	}
}

func (s *InMemorySessionManager) maybeSendMessage(ctx context.Context, msg XMessage, sess *Session) {
	switch sess.SendMessage(msg) {
	case SessSendClosed:
		s.logger.WarnContext(ctx, "can't send notification because the user's session is closed", "recipient", sess.ScreenName, "message", msg)
	case SessSendTimeout:
		s.logger.WarnContext(ctx, "can't send notification because of send timeout", "recipient", sess.ScreenName, "message", msg)
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

func (s *InMemorySessionManager) BroadcastExcept(ctx context.Context, except *Session, msg XMessage) {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	for _, sess := range s.store {
		if sess == except {
			continue
		}
		go s.maybeSendMessage(ctx, msg, sess)
	}
}

func (s *InMemorySessionManager) Retrieve(ID string) (*Session, bool) {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	sess, found := s.store[ID]
	return sess, found
}

func (s *InMemorySessionManager) RetrieveByScreenName(screenName string) (*Session, error) {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	for _, sess := range s.store {
		if screenName == sess.ScreenName {
			return sess, nil
		}
	}
	return nil, fmt.Errorf("%w: %s", ErrSessNotFound, screenName)
}

func (s *InMemorySessionManager) retrieveByScreenNames(screenNames []string) []*Session {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	var ret []*Session
	for _, sn := range screenNames {
		for _, sess := range s.store {
			if sn == sess.ScreenName {
				ret = append(ret, sess)
			}
		}
	}
	return ret
}

func (s *InMemorySessionManager) SendToScreenName(ctx context.Context, screenName string, msg XMessage) {
	sess, err := s.RetrieveByScreenName(screenName)
	if err != nil {
		s.logger.WarnContext(ctx, "can't send notification because user is not online", "recipient", screenName, "message", msg)
		return
	}
	go s.maybeSendMessage(ctx, msg, sess)
}

func (s *InMemorySessionManager) BroadcastToScreenNames(ctx context.Context, screenNames []string, msg XMessage) {
	for _, sess := range s.retrieveByScreenNames(screenNames) {
		go s.maybeSendMessage(ctx, msg, sess)
	}
}

func makeSession() *Session {
	return &Session{
		msgCh:       make(chan XMessage, 1),
		stopCh:      make(chan struct{}),
		sendTimeout: sendTimeout,
		SignonTime:  time.Now(),
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
		if screenName == sess.ScreenName {
			sess.Close()
			delete(s.store, sess.ID)
			break
		}
	}

	sess := makeSession()
	sess.ID = sessID
	sess.ScreenName = screenName
	s.store[sess.ID] = sess
	return sess
}

func (s *InMemorySessionManager) Remove(sess *Session) {
	s.mapMutex.Lock()
	defer s.mapMutex.Unlock()
	delete(s.store, sess.ID)
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
