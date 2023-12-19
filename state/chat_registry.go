package state

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mkaminski/goaim/oscar"
)

type ChatRegistry struct {
	chatRoomStore map[string]ChatRoom
	smStore       map[string]any
	mapMutex      sync.RWMutex
}

func NewChatRegistry() *ChatRegistry {
	return &ChatRegistry{
		chatRoomStore: make(map[string]ChatRoom),
		smStore:       make(map[string]any),
	}
}

func (c *ChatRegistry) Register(room ChatRoom, sessionManager any) {
	c.mapMutex.Lock()
	defer c.mapMutex.Unlock()
	c.chatRoomStore[room.Cookie] = room
	c.smStore[room.Cookie] = sessionManager
}

var ErrChatRoomNotFound = errors.New("chat room not found")

func (c *ChatRegistry) Retrieve(cookie string) (ChatRoom, any, error) {
	c.mapMutex.RLock()
	defer c.mapMutex.RUnlock()
	chatRoom, found := c.chatRoomStore[cookie]
	if !found {
		return ChatRoom{}, nil, fmt.Errorf("%w cookie: %s", ErrChatRoomNotFound, cookie)
	}
	sessionManager, found := c.smStore[cookie]
	if !found {
		panic("unable to find session manager for chat")
	}
	return chatRoom, sessionManager, nil
}

func (c *ChatRegistry) RemoveRoom(cookie string) {
	c.mapMutex.Lock()
	defer c.mapMutex.Unlock()
	delete(c.chatRoomStore, cookie)
	delete(c.smStore, cookie)
}

type ChatRoom struct {
	CreateTime     time.Time
	DetailLevel    uint8
	Exchange       uint16
	Cookie         string
	InstanceNumber uint16
	Name           string
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

// NewChatRoom creates new state.ChatRoom objects
func NewChatRoom() ChatRoom {
	return ChatRoom{
		Cookie:     uuid.New().String(),
		CreateTime: time.Now(),
	}
}
