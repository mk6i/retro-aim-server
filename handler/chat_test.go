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
		inputSNAC oscar.SNACMessage
		// expectSNACToParticipants is the message the server broadcast to chat
		// room participants (except the sender)
		expectSNACToParticipants oscar.SNACMessage
		expectOutput             *oscar.SNACMessage
	}{
		{
			name:        "send chat room message, expect acknowledgement to sender client",
			userSession: newTestSession("user_sending_chat_msg", sessOptCannedSignonTime),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost{
					Cookie:  1234,
					Channel: 14,
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								Tag:   oscar.ChatTLVPublicWhisperFlag,
								Value: []byte{},
							},
							{
								Tag:   oscar.ChatTLVEnableReflectionFlag,
								Value: []byte{},
							},
						},
					},
				},
			},
			expectSNACToParticipants: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Chat,
					SubGroup:  oscar.ChatChannelMsgToClient,
				},
				Body: oscar.SNAC_0x0E_0x06_ChatChannelMsgToClient{
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
			expectOutput: &oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Chat,
					SubGroup:  oscar.ChatChannelMsgToClient,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x0E_0x06_ChatChannelMsgToClient{
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
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost{
					Cookie:  1234,
					Channel: 14,
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								Tag:   oscar.ChatTLVPublicWhisperFlag,
								Value: []byte{},
							},
						},
					},
				},
			},
			expectSNACToParticipants: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Chat,
					SubGroup:  oscar.ChatChannelMsgToClient,
				},
				Body: oscar.SNAC_0x0E_0x06_ChatChannelMsgToClient{
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
			expectOutput: &oscar.SNACMessage{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chatID := "the-chat-id"

			chatSessMgr := newMockChatMessageRelayer(t)
			chatSessMgr.EXPECT().
				RelayToAllExcept(mock.Anything, tc.userSession, tc.expectSNACToParticipants)

			svc := ChatService{
				chatRegistry: state.NewChatRegistry(),
			}
			svc.chatRegistry.Register(state.ChatRoom{Cookie: chatID}, chatSessMgr)

			outputSNAC, err := svc.ChannelMsgToHostHandler(context.Background(), tc.userSession, chatID,
				tc.inputSNAC.Frame, tc.inputSNAC.Body.(oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost))
			assert.NoError(t, err)

			if tc.expectOutput.Frame == (oscar.SNACFrame{}) {
				return // handler doesn't return response
			}

			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}
