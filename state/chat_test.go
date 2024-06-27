package state

import (
	"testing"

	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
)

func TestChatRoom_TLVList(t *testing.T) {
	room := NewChatRoom()
	room.Name = "chat-room-name"

	have := room.TLVList()
	want := []wire.TLV{
		wire.NewTLV(wire.ChatRoomTLVFlags, uint16(15)),
		wire.NewTLV(wire.ChatRoomTLVCreateTime, uint32(room.CreateTime.Unix())),
		wire.NewTLV(wire.ChatRoomTLVMaxMsgLen, uint16(1024)),
		wire.NewTLV(wire.ChatRoomTLVMaxOccupancy, uint16(100)),
		wire.NewTLV(wire.ChatRoomTLVNavCreatePerms, uint8(2)),
		wire.NewTLV(wire.ChatRoomTLVFullyQualifiedName, room.Name),
		wire.NewTLV(wire.ChatRoomTLVRoomName, room.Name),
	}

	assert.Equal(t, want, have)
}
