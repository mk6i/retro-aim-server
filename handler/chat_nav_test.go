package handler

import (
	"context"
	"testing"
	"time"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
	"github.com/stretchr/testify/assert"
)

func TestSendAndReceiveCreateRoom(t *testing.T) {
	//
	// build dependencies
	//
	userSess := newTestSession("user-screen-name", sessOptCannedID)

	cr := state.NewChatRegistry()

	sm := newMockChatSessionManager(t)
	sm.EXPECT().NewSessionWithSN(userSess.ID(), userSess.ScreenName()).
		Return(&state.Session{})

	//
	// send input SNAC
	//
	inputSNAC := oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
		Exchange:       1,
		Cookie:         "create", // actual canned value sent by AIM client
		InstanceNumber: 2,
		DetailLevel:    3,
		TLVBlock: oscar.TLVBlock{
			TLVList: oscar.TLVList{
				oscar.NewTLV(oscar.ChatTLVRoomName, "the-chat-room-name"),
			},
		},
	}
	svc := ChatNavService{
		chatRegistry: cr,
		newChatRoom: func() state.ChatRoom {
			return state.ChatRoom{
				Cookie:     "dummy-cookie",
				CreateTime: time.UnixMilli(0),
			}
		},
		newChatSessMgr: func() ChatSessionManager {
			return sm
		},
	}
	outputSNAC, err := svc.CreateRoomHandler(context.Background(), userSess, inputSNAC)
	assert.NoError(t, err)

	//
	// verify chat room created by handler
	//
	expectChatRoom := state.ChatRoom{
		Cookie:         "dummy-cookie",
		CreateTime:     time.UnixMilli(0),
		DetailLevel:    3,
		Exchange:       1,
		InstanceNumber: 2,
		Name:           "the-chat-room-name",
	}
	chatRoom, _, err := cr.Retrieve("dummy-cookie")
	assert.NoError(t, err)
	assert.Equal(t, expectChatRoom, chatRoom)

	//
	// send input SNAC
	//
	expectSNAC := oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.ChatNav,
			SubGroup:  oscar.ChatNavNavInfo,
		},
		Body: oscar.SNAC_0x0D_0x09_ChatNavNavInfo{
			TLVRestBlock: oscar.TLVRestBlock{
				TLVList: oscar.TLVList{
					oscar.NewTLV(
						oscar.ChatNavTLVRoomInfo,
						oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
							Exchange:       chatRoom.Exchange,
							Cookie:         chatRoom.Cookie,
							InstanceNumber: chatRoom.InstanceNumber,
							DetailLevel:    chatRoom.DetailLevel,
							TLVBlock: oscar.TLVBlock{
								TLVList: chatRoom.TLVList(),
							},
						},
					),
				},
			},
		},
	}

	assert.Equal(t, expectSNAC, outputSNAC)
}
