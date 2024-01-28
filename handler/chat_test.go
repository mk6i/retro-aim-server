package handler

import (
	"context"
	"testing"

	"github.com/mk6i/retro-aim-server/oscar"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestChatService_ChannelMsgToHostHandler(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user sending the chat message
		userSession *state.Session
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC oscar.SNACMessage
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// expectSNACToParticipants is the message the server broadcast to chat
		// room participants (except the sender)
		expectSNACToParticipants oscar.SNACMessage
		expectOutput             *oscar.SNACMessage
		wantErr                  error
	}{
		{
			name: "send chat room message, expect acknowledgement to sender client",
			userSession: newTestSession("user_sending_chat_msg", sessOptCannedSignonTime,
				sessOptChatID("the-chat-id")),
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
			mockParams: mockParams{
				chatRegistryParams: chatRegistryParams{
					chatRegistryRetrieveParams: chatRegistryRetrieveParams{
						chatID: "the-chat-id",
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
			name: "send chat room message, don't expect acknowledgement to sender client",
			userSession: newTestSession("user_sending_chat_msg", sessOptCannedSignonTime,
				sessOptChatID("the-chat-id")),
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
			mockParams: mockParams{
				chatRegistryParams: chatRegistryParams{
					chatRegistryRetrieveParams: chatRegistryRetrieveParams{
						chatID: "the-chat-id",
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
		},
		{
			name: "send chat room message, fail due to missing chat room",
			userSession: newTestSession("user_sending_chat_msg", sessOptCannedSignonTime,
				sessOptChatID("the-chat-id")),
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
			mockParams: mockParams{
				chatRegistryParams: chatRegistryParams{
					chatRegistryRetrieveParams: chatRegistryRetrieveParams{
						chatID: "the-chat-id",
						err:    state.ErrChatRoomNotFound,
					},
				},
			},
			wantErr: state.ErrChatRoomNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chatSessMgr := newMockChatMessageRelayer(t)
			if tc.mockParams.chatRegistryRetrieveParams.err == nil {
				chatSessMgr.EXPECT().
					RelayToAllExcept(mock.Anything, tc.userSession, tc.expectSNACToParticipants)
			}

			chatRegistry := newMockChatRegistry(t)
			chatRegistry.EXPECT().
				Retrieve(tc.mockParams.chatRegistryRetrieveParams.chatID).
				Return(state.ChatRoom{}, chatSessMgr, tc.mockParams.chatRegistryRetrieveParams.err)

			svc := NewChatService(chatRegistry)
			outputSNAC, err := svc.ChannelMsgToHostHandler(context.Background(), tc.userSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost))
			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}
