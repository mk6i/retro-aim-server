package handler

import (
	"context"
	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/server"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSendAndReceiveCreateRoom(t *testing.T) {
	//
	// build dependencies
	//
	userSess := newTestSession("user-screen-name", sessOptCannedID)

	cr := server.NewChatRegistry()

	sm := server.NewMockChatSessionManager(t)
	sm.EXPECT().NewSessionWithSN(userSess.ID(), userSess.ScreenName()).
		Return(&server.Session{})

	chatSessMgrFactory := func() server.ChatSessionManager {
		return sm
	}

	newChatRoom := func() server.ChatRoom {
		return server.ChatRoom{
			Cookie:     "dummy-cookie",
			CreateTime: time.UnixMilli(0),
		}
	}

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
		cr: cr,
	}
	outputSNAC, err := svc.CreateRoomHandler(context.Background(), userSess, newChatRoom, chatSessMgrFactory, inputSNAC)
	assert.NoError(t, err)

	//
	// verify chat room created by handler
	//
	expectChatRoom := server.ChatRoom{
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
	expectSNAC := oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.CHAT_NAV,
			SubGroup:  oscar.ChatNavNavInfo,
		},
		SnacOut: oscar.SNAC_0x0D_0x09_ChatNavNavInfo{
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
