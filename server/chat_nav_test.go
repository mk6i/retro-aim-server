package server

import (
	"bytes"
	"testing"
	"time"

	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
)

func TestSendAndReceiveCreateRoom(t *testing.T) {
	//
	// build dependencies
	//
	userSess := newTestSession(Session{
		ID:         "sess-id",
		ScreenName: "user-screen-name",
	})

	cr := NewChatRegistry()

	sm := NewMockSessionManager(t)
	sm.EXPECT().NewSessionWithSN(userSess.ID, userSess.ScreenName).
		Return(&Session{})

	crf := func() ChatRoom {
		return ChatRoom{
			Cookie:         "dummy-cookie",
			CreateTime:     time.UnixMilli(0),
			SessionManager: sm,
		}
	}

	//
	// send input SNAC
	//
	inputSNACFrame := oscar.SnacFrame{
		FoodGroup: CHAT_NAV,
		SubGroup:  ChatNavCreateRoom,
	}
	inputSNAC := oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
		Exchange:       1,
		Cookie:         "create", // actual canned value sent by AIM client
		InstanceNumber: 2,
		DetailLevel:    3,
		TLVBlock: oscar.TLVBlock{
			TLVList: oscar.TLVList{
				{
					TType: oscar.ChatTLVRoomName,
					Val:   "the-chat-room-name",
				},
			},
		},
	}
	input := &bytes.Buffer{}
	assert.NoError(t, oscar.Marshal(inputSNAC, input))

	var seq uint32
	output := &bytes.Buffer{}
	assert.NoError(t, SendAndReceiveCreateRoom(userSess, cr, crf, inputSNACFrame, input, output, &seq))

	//
	// verify chat room created by handler
	//
	expectChatRoom := ChatRoom{
		SessionManager: sm,
		Cookie:         "dummy-cookie",
		CreateTime:     time.UnixMilli(0),
		DetailLevel:    3,
		Exchange:       1,
		InstanceNumber: 2,
		Name:           "the-chat-room-name",
	}
	chatRoom, err := cr.Retrieve("dummy-cookie")
	assert.NoError(t, err)
	assert.Equal(t, expectChatRoom, chatRoom)

	//
	// verify SNAC frame
	//
	expectSNACFrame := oscar.SnacFrame{
		FoodGroup: CHAT_NAV,
		SubGroup:  ChatNavNavInfo,
	}
	flap := oscar.FlapFrame{}
	assert.NoError(t, oscar.Unmarshal(&flap, output))
	snacFrame := oscar.SnacFrame{}
	assert.NoError(t, oscar.Unmarshal(&snacFrame, output))
	assert.Equal(t, expectSNACFrame, snacFrame)

	//
	// verify SNAC body
	//
	expectSNAC := oscar.SNAC_0x0D_0x09_ChatNavNavInfo{
		TLVRestBlock: oscar.TLVRestBlock{
			TLVList: oscar.TLVList{
				{
					TType: oscar.ChatNavTLVRoomInfo,
					Val: oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
						Exchange:       chatRoom.Exchange,
						Cookie:         chatRoom.Cookie,
						InstanceNumber: chatRoom.InstanceNumber,
						DetailLevel:    chatRoom.DetailLevel,
						TLVBlock: oscar.TLVBlock{
							TLVList: chatRoom.TLVList(),
						},
					},
				},
			},
		},
	}
	assert.NoError(t, expectSNAC.SerializeInPlace())
	outputSNAC := oscar.SNAC_0x0D_0x09_ChatNavNavInfo{}
	assert.NoError(t, oscar.Unmarshal(&outputSNAC, output))
	assert.Equal(t, expectSNAC, outputSNAC)

	assert.Equalf(t, 0, output.Len(), "the rest of the buffer is unread")
}
