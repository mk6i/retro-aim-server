package state

import (
	"testing"

	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
)

func TestChatRegistry_RegisterAndReceive(t *testing.T) {
	type registration struct {
		room  ChatRoom
		value any
	}
	tests := []struct {
		name            string
		givenRegistered []registration
		lookupCookie    string
		wantRegistered  registration
		wantErr         error
	}{
		{
			name: "chat room and value found",
			givenRegistered: []registration{
				{
					room:  ChatRoom{Cookie: "cookie1"},
					value: "value1",
				},
				{
					room:  ChatRoom{Cookie: "cookie2"},
					value: "value2",
				},
			},
			lookupCookie: "cookie2",
			wantRegistered: registration{
				room:  ChatRoom{Cookie: "cookie2"},
				value: "value2",
			},
		},
		{
			name: "chat room and value not found",
			givenRegistered: []registration{
				{
					room:  ChatRoom{Cookie: "cookie1"},
					value: "value1",
				},
				{
					room:  ChatRoom{Cookie: "cookie2"},
					value: "value2",
				},
			},
			lookupCookie: "cookie3",
			wantErr:      ErrChatRoomNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatRegistry := NewChatRegistry()
			for _, r := range tt.givenRegistered {
				chatRegistry.Register(r.room, r.value)
			}

			room, value, err := chatRegistry.Retrieve(tt.lookupCookie)

			assert.Equal(t, tt.wantRegistered.room, room)
			assert.Equal(t, tt.wantRegistered.value, value)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestChatRegistry_RegisterAndRemove(t *testing.T) {
	type registration struct {
		room  ChatRoom
		value any
	}
	tests := []struct {
		name            string
		givenRegistered []registration
		removeCookie    string
		wantRegistered  []registration
		wantErr         error
	}{
		{
			name: "chat room and value removed",
			givenRegistered: []registration{
				{
					room:  ChatRoom{Cookie: "cookie1"},
					value: "value1",
				},
				{
					room:  ChatRoom{Cookie: "cookie2"},
					value: "value2",
				},
			},
			removeCookie: "cookie2",
			wantRegistered: []registration{
				{
					room:  ChatRoom{Cookie: "cookie1"},
					value: "value1",
				},
			},
		},
		{
			name: "no chat room and value removed",
			givenRegistered: []registration{
				{
					room:  ChatRoom{Cookie: "cookie1"},
					value: "value1",
				},
				{
					room:  ChatRoom{Cookie: "cookie2"},
					value: "value2",
				},
			},
			removeCookie: "cookie3",
			wantRegistered: []registration{
				{
					room:  ChatRoom{Cookie: "cookie1"},
					value: "value1",
				},
				{
					room:  ChatRoom{Cookie: "cookie2"},
					value: "value2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatRegistry := NewChatRegistry()
			for _, r := range tt.givenRegistered {
				chatRegistry.Register(r.room, r.value)
			}

			chatRegistry.Remove(tt.removeCookie)

			for _, r := range tt.wantRegistered {
				room, value, err := chatRegistry.Retrieve(r.room.Cookie)
				assert.Equal(t, r.room, room)
				assert.Equal(t, r.value, value)
				assert.NoError(t, err)
			}
		})
	}
}

func TestChatRoom_TLVList(t *testing.T) {
	room := NewChatRoom()
	room.Name = "chat-room-name"

	have := room.TLVList()
	want := []wire.TLV{
		wire.NewTLV(wire.ChatNavTLVFlags, uint16(15)),
		wire.NewTLV(wire.ChatNavCreateTime, uint32(room.CreateTime.Unix())),
		wire.NewTLV(wire.ChatNavTLVMaxMsgLen, uint16(1024)),
		wire.NewTLV(wire.ChatNavTLVMaxOccupancy, uint16(100)),
		wire.NewTLV(wire.ChatNavTLVCreatePerms, uint8(2)),
		wire.NewTLV(wire.ChatNavTLVFullyQualifiedName, room.Name),
		wire.NewTLV(wire.ChatNavTLVRoomName, room.Name),
	}

	assert.Equal(t, want, have)
}
