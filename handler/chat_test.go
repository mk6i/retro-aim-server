package handler

import (
	"context"
	"testing"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendAndReceiveChatChannelMsgToHost(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user sending the chat message
		userSession *state.Session
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost
		// expectSNACToParticipants is the message the server broadcast to chat
		// room participants (except the sender)
		expectSNACToParticipants oscar.XMessage
		expectOutput             *oscar.XMessage
	}{
		{
			name:        "send chat room message, expect acknowledgement to sender client",
			userSession: newTestSession("user_sending_chat_msg", sessOptCannedSignonTime),
			inputSNAC: oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost{
				Cookie:  1234,
				Channel: 14,
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						{
							TType: oscar.ChatTLVPublicWhisperFlag,
							Val:   []byte{},
						},
						{
							TType: oscar.ChatTLVEnableReflectionFlag,
							Val:   []byte{},
						},
					},
				},
			},
			expectSNACToParticipants: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.CHAT,
					SubGroup:  oscar.ChatChannelMsgToClient,
				},
				SnacOut: oscar.SNAC_0x0E_0x06_ChatChannelMsgToClient{
					Cookie:  1234,
					Channel: 14,
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(oscar.ChatTLVPublicWhisperFlag, []byte{}),
							oscar.NewTLV(oscar.ChatTLVEnableReflectionFlag, []byte{}),
							oscar.NewTLV(oscar.ChatTLVSenderInformation,
								newTestSession("user_sending_chat_msg", sessOptCannedSignonTime).TLVUserInfo()),
						},
					},
				},
			},
			expectOutput: &oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.CHAT,
					SubGroup:  oscar.ChatChannelMsgToClient,
				},
				SnacOut: oscar.SNAC_0x0E_0x06_ChatChannelMsgToClient{
					Cookie:  1234,
					Channel: 14,
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(oscar.ChatTLVPublicWhisperFlag, []byte{}),
							oscar.NewTLV(oscar.ChatTLVEnableReflectionFlag, []byte{}),
							oscar.NewTLV(oscar.ChatTLVSenderInformation, newTestSession("user_sending_chat_msg", sessOptCannedSignonTime).TLVUserInfo()),
						},
					},
				},
			},
		},
		{
			name:        "send chat room message, don't expect acknowledgement to sender client",
			userSession: newTestSession("user_sending_chat_msg", sessOptCannedSignonTime),
			inputSNAC: oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost{
				Cookie:  1234,
				Channel: 14,
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						{
							TType: oscar.ChatTLVPublicWhisperFlag,
							Val:   []byte{},
						},
					},
				},
			},
			expectSNACToParticipants: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.CHAT,
					SubGroup:  oscar.ChatChannelMsgToClient,
				},
				SnacOut: oscar.SNAC_0x0E_0x06_ChatChannelMsgToClient{
					Cookie:  1234,
					Channel: 14,
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(oscar.ChatTLVPublicWhisperFlag, []byte{}),
							oscar.NewTLV(oscar.ChatTLVSenderInformation,
								newTestSession("user_sending_chat_msg", sessOptCannedSignonTime).TLVUserInfo()),
						},
					},
				},
			},
			expectOutput: &oscar.XMessage{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chatID := "the-chat-id"

			chatSessMgr := newMockChatSessionManager(t)
			chatSessMgr.EXPECT().
				BroadcastExcept(mock.Anything, tc.userSession, tc.expectSNACToParticipants)

			svc := ChatService{
				chatRegistry: state.NewChatRegistry(),
			}
			svc.chatRegistry.Register(state.ChatRoom{Cookie: chatID}, chatSessMgr)

			outputSNAC, err := svc.ChannelMsgToHostHandler(context.Background(), tc.userSession, chatID, tc.inputSNAC)
			assert.NoError(t, err)

			if tc.expectOutput.SnacFrame == (oscar.SnacFrame{}) {
				return // handler doesn't return response
			}

			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}
