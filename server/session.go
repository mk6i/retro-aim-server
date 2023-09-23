package server

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/mkaminski/goaim/oscar"
	"sync"
	"time"
)

var errSessNotFound = errors.New("session was not found")

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

func (s *Session) GetUserInfo() []oscar.TLV {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()

	// sign-in timestamp
	tlvs := []oscar.TLV{
		{
			TType: 0x03,
			Val:   uint32(s.SignonTime.Unix()),
		},
	}

	// away message status
	userFlags := oscar.TLV{
		TType: 0x01,
		Val:   uint16(0x0010), // AIM client
	}
	if s.AwayMessage != "" {
		userFlags.Val = userFlags.Val.(uint16) | uint16(0x0020)
	}
	tlvs = append(tlvs, userFlags)

	// invisibility status
	status := oscar.TLV{
		TType: 0x06,
		Val:   uint16(0x0000),
	}
	if s.invisible {
		status.Val = status.Val.(uint16) | uint16(0x0100)
	}
	tlvs = append(tlvs, status)

	// idle status
	idle := oscar.TLV{
		TType: 0x04,
		Val:   uint16(0),
	}
	if s.idle {
		idle.Val = uint16(time.Now().Sub(s.idleTime).Seconds())
	}
	tlvs = append(tlvs, idle)

	// capabilities
	caps := oscar.TLV{
		TType: 0x0D,
		Val:   []byte{},
	}
	// chat capability
	caps.Val = append(caps.Val.([]byte), CapChat...)
	tlvs = append(tlvs, caps)

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

type SessionManager struct {
	store    map[string]*Session
	mapMutex sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		store: make(map[string]*Session),
	}
}

func (s *SessionManager) Broadcast(msg XMessage) {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	for _, sess := range s.store {
		go s.maybeSendMessage(msg, sess)
	}
}

func (s *SessionManager) maybeSendMessage(msg XMessage, sess *Session) {
	switch sess.SendMessage(msg) {
	case SessSendClosed:
		fmt.Printf("message to %s was blocked, removing session\n", sess.ScreenName)
	case SessSendTimeout:
		fmt.Printf("message to %s timed out\n", sess.ScreenName)
		sess.Close()
	}
}

func (s *SessionManager) Empty() bool {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	return len(s.store) == 0
}

func (s *SessionManager) All() []*Session {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	var sessions []*Session
	for _, sess := range s.store {
		sessions = append(sessions, sess)
	}
	return sessions
}

func (s *SessionManager) BroadcastExcept(except *Session, msg XMessage) {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	for _, sess := range s.store {
		if sess == except {
			continue
		}
		go s.maybeSendMessage(msg, sess)
	}
}

func (s *SessionManager) Retrieve(ID string) (*Session, bool) {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	sess, found := s.store[ID]
	return sess, found
}

func (s *SessionManager) RetrieveByScreenName(screenName string) (*Session, error) {
	s.mapMutex.RLock()
	defer s.mapMutex.RUnlock()
	for _, sess := range s.store {
		if screenName == sess.ScreenName {
			return sess, nil
		}
	}
	return nil, fmt.Errorf("%w: %s", errSessNotFound, screenName)
}

func (s *SessionManager) RetrieveByScreenNames(screenNames []string) []*Session {
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

func (s *SessionManager) SendToScreenName(screenName string, msg XMessage) {
	sess, err := s.RetrieveByScreenName(screenName)
	if err != nil {
		fmt.Printf("error sending to screen name: %s\n", screenName)
		return
	}
	go s.maybeSendMessage(msg, sess)
}

func (s *SessionManager) BroadcastToScreenNames(screenNames []string, msg XMessage) {
	for _, sess := range s.RetrieveByScreenNames(screenNames) {
		go s.maybeSendMessage(msg, sess)
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

func (s *SessionManager) NewSession() (*Session, error) {
	s.mapMutex.Lock()
	defer s.mapMutex.Unlock()
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	sess := makeSession()
	sess.ID = id.String()
	s.store[sess.ID] = sess
	return sess, nil
}

func (s *SessionManager) NewSessionWithSN(sessID string, screenName string) *Session {
	s.mapMutex.Lock()
	defer s.mapMutex.Unlock()
	sess := makeSession()
	sess.ID = sessID
	sess.ScreenName = screenName
	s.store[sess.ID] = sess
	return sess
}

func (s *SessionManager) Remove(sess *Session) {
	s.mapMutex.Lock()
	defer s.mapMutex.Unlock()
	delete(s.store, sess.ID)
}

type ChatRoom struct {
	ID             string
	SessionManager *SessionManager
	CreateTime     time.Time
	Name           string
}

func (c ChatRoom) TLVList() []oscar.TLV {
	return []oscar.TLV{
		{
			TType: 0x00c9,
			Val:   uint16(15),
		},
		{
			TType: 0x00ca,
			Val:   uint32(c.CreateTime.Unix()),
		},
		{
			TType: 0x00d1,
			Val:   uint16(1024),
		},
		{
			TType: 0x00d2,
			Val:   uint16(100),
		},
		{
			TType: 0x00d5,
			Val:   uint8(2),
		},
		{
			TType: 0x006a,
			Val:   c.Name,
		},
		{
			TType: 0x00d3,
			Val:   c.Name,
		},
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
	c.store[room.ID] = room
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
	if found && room.SessionManager.Empty() {
		delete(c.store, chatID)
	}
}
