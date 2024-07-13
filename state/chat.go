package state

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/mk6i/retro-aim-server/wire"
)

const (
	// PrivateExchange is the ID of the exchange that hosts non-public created
	// by users.
	PrivateExchange uint16 = 4
	// PublicExchange is the ID of the exchange that hosts public chat rooms
	// created by the server operator exclusively.
	PublicExchange uint16 = 5
)

// ErrChatRoomNotFound indicates that a chat room lookup failed.
var (
	ErrChatRoomNotFound = errors.New("chat room not found")
	ErrDupChatRoom      = errors.New("chat room already exists")
)

// ChatRoom is the representation of a chat room's metadata.
type ChatRoom struct {
	// Cookie is the unique chat room identifier.
	Cookie string
	// CreateTime indicates when the chat room was created.
	CreateTime time.Time
	// Creator is the screen name of the user who created the chat room.
	Creator IdentScreenName
	// Exchange indicates which exchange the chatroom belongs to. Typically, a canned value.
	Exchange uint16
	// InstanceNumber indicates which instance chatroom exists in. Typically, a canned value.
	InstanceNumber uint16
	// Name is the name of the chat room.
	Name string
}

// URL creates a URL that can be used to join a chat room.
func (c ChatRoom) URL() *url.URL {
	v := url.Values{}
	v.Set("roomname", c.Name)
	v.Set("exchange", fmt.Sprintf("%d", c.Exchange))

	return &url.URL{
		Scheme: "aim",
		Opaque: "gochat?" + v.Encode(),
	}
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
		wire.NewTLV(wire.ChatRoomTLVMaxMsgVisLen, uint16(1024)),
	}
}

// NewChatRoom creates new state.ChatRoom objects
func NewChatRoom() ChatRoom {
	return ChatRoom{
		CreateTime: time.Now().UTC(),
	}
}
