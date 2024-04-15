package state

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/mk6i/retro-aim-server/wire"

	"github.com/google/uuid"
)

// ErrChatRoomNotFound indicates that a chat room lookup failed.
var ErrChatRoomNotFound = errors.New("chat room not found")

// ChatRegistry keeps track of chat rooms. A ChatRegistry is safe for
// concurrent use by multiple goroutines.
type ChatRegistry struct {
	roomStore  map[string]ChatRoom // association of cookie identifier->chat room
	valueStore map[string]any      // association of cookie identifier->value
	mutex      sync.RWMutex        // ensures thread-safe read-write access to stores
}

// NewChatRegistry creates a new instance of ChatRegistry
func NewChatRegistry() *ChatRegistry {
	return &ChatRegistry{
		roomStore:  make(map[string]ChatRoom),
		valueStore: make(map[string]any),
	}
}

// Register adds a chat room to the registry and associates an arbitrary value
// with the room.
func (c *ChatRegistry) Register(chatRoom ChatRoom, value any) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.roomStore[chatRoom.Cookie] = chatRoom
	c.valueStore[chatRoom.Cookie] = value
}

// Retrieve retrieves a chat room and the arbitrary value associated with it.
// Returns ErrChatRoomNotFound if the room is not registered.
func (c *ChatRegistry) Retrieve(cookie string) (ChatRoom, any, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	chatRoom, found := c.roomStore[cookie]
	if !found {
		return ChatRoom{}, nil, fmt.Errorf("%w cookie: %s", ErrChatRoomNotFound, cookie)
	}
	value, found := c.valueStore[cookie]
	if !found {
		panic("unable to find value for chat room")
	}
	return chatRoom, value, nil
}

// Remove removes a chat room and the arbitrary value associated with it.
func (c *ChatRegistry) Remove(cookie string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.roomStore, cookie)
	delete(c.valueStore, cookie)
}

// ChatRoom is the representation of a chat room's metadata.
type ChatRoom struct {
	// Cookie is the unique chat room identifier.
	Cookie string
	// CreateTime indicates when the chat room was created.
	CreateTime time.Time
	// DetailLevel is the detail level of the chat room.  Unclear what this value means.
	DetailLevel uint8
	// Exchange indicates which exchange the chatroom belongs to. Typically, a canned value.
	Exchange uint16
	// InstanceNumber indicates which instance chatroom exists in. Typically, a canned value.
	InstanceNumber uint16
	// Name is the name of the chat room.
	Name string
}

// TLVList returns a TLV list of chat room metadata.
func (c ChatRoom) TLVList() []wire.TLV {
	return []wire.TLV{
		// From protocols/oscar/family_chatnav.c in lib purple, these are the
		// room creation flags:
		// - 1 Evilable
		// - 2 Nav Only
		// - 4 Instancing Allowed
		// - 8 Occupant Peek Allowed
		// It's unclear what effect they actually have.
		wire.NewTLV(wire.ChatRoomTLVFlags, uint16(15)),
		wire.NewTLV(wire.ChatRoomTLVCreateTime, uint32(c.CreateTime.Unix())),
		wire.NewTLV(wire.ChatRoomTLVMaxMsgLen, uint16(1024)),
		wire.NewTLV(wire.ChatRoomTLVMaxOccupancy, uint16(100)),
		// From protocols/oscar/family_chatnav.c in lib purple, these are the
		// room creation permission values:
		// - 0  creation not allowed
		// - 1  room creation allowed
		// - 2  exchange creation allowed
		// It's unclear what effect they actually have.
		wire.NewTLV(wire.ChatRoomTLVNavCreatePerms, uint8(2)),
		wire.NewTLV(wire.ChatRoomTLVFullyQualifiedName, c.Name),
		wire.NewTLV(wire.ChatRoomTLVRoomName, c.Name),
	}
}

// NewChatRoom creates new state.ChatRoom objects
func NewChatRoom() ChatRoom {
	return ChatRoom{
		Cookie:     uuid.New().String(),
		CreateTime: time.Now(),
	}
}
