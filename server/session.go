package server

import (
	"errors"
	"sync"
	"time"

	"github.com/mkaminski/goaim/oscar"
)

var ErrSignedOff = errors.New("user signed off")

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

type ChatRegistry struct {
	chatRoomStore map[string]ChatRoom
	smStore       map[string]ChatSessionManager
	mapMutex      sync.RWMutex
}

func NewChatRegistry() *ChatRegistry {
	return &ChatRegistry{
		chatRoomStore: make(map[string]ChatRoom),
		smStore:       make(map[string]ChatSessionManager),
	}
}

func (c *ChatRegistry) Register(room ChatRoom, sm ChatSessionManager) {
	c.mapMutex.Lock()
	defer c.mapMutex.Unlock()
	c.chatRoomStore[room.Cookie] = room
	c.smStore[room.Cookie] = sm
}

func (c *ChatRegistry) Retrieve(chatID string) (ChatRoom, ChatSessionManager, error) {
	c.mapMutex.RLock()
	defer c.mapMutex.RUnlock()
	cr, found := c.chatRoomStore[chatID]
	if !found {
		return ChatRoom{}, nil, errors.New("unable to find chat room")
	}
	sm, found := c.smStore[chatID]
	if !found {
		panic("unable to find session manager for chat")
	}
	return cr, sm, nil
}

func (c *ChatRegistry) MaybeRemoveRoom(chatID string) {
	c.mapMutex.Lock()
	defer c.mapMutex.Unlock()
	sm, found := c.smStore[chatID]
	if found && sm.Empty() {
		delete(c.chatRoomStore, chatID)
		delete(c.smStore, chatID)
	}
}
