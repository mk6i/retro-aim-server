package foodgroup

import (
	"context"
	"math"
	"testing"

	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestChatService_ChannelMsgToHost(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user sending the chat message
		userSession *state.Session
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC wire.SNACMessage
		// mockParams is the list of params sent to mocks that satisfy this
		// method's dependencies
		mockParams mockParams
		// expectSNACToParticipants is the message the server broadcast to chat
		// room participants (except the sender)
		expectSNACToParticipants wire.SNACMessage
		expectOutput             *wire.SNACMessage
		wantErr                  error
	}{
		{
			name: "send chat room message, expect acknowledgement to sender client",
			userSession: newTestSession("user_sending_chat_msg", sessOptCannedSignonTime,
				sessOptChatRoomCookie("the-chat-cookie")),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x0E_0x05_ChatChannelMsgToHost{
					Cookie:  1234,
					Channel: 14,
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							{
								Tag:   wire.ChatTLVPublicWhisperFlag,
								Value: []byte{},
							},
							{
								Tag:   wire.ChatTLVEnableReflectionFlag,
								Value: []byte{},
							},
							{
								Tag:   wire.ChatTLVMessageInformation,
								Value: []byte{},
							},
						},
					},
				},
			},
			mockParams: mockParams{
				chatMessageRelayerParams: chatMessageRelayerParams{
					chatRelayToAllExceptParams: chatRelayToAllExceptParams{
						{
							screenName: state.NewIdentScreenName("user_sending_chat_msg"),
							cookie:     "the-chat-cookie",
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Chat,
									SubGroup:  wire.ChatChannelMsgToClient,
								},
								Body: wire.SNAC_0x0E_0x06_ChatChannelMsgToClient{
									Cookie:  1234,
									Channel: 14,
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLVBE(wire.ChatTLVSenderInformation,
												newTestSession("user_sending_chat_msg", sessOptCannedSignonTime).TLVUserInfo()),
											wire.NewTLVBE(wire.ChatTLVPublicWhisperFlag, []byte{}),
											wire.NewTLVBE(wire.ChatTLVMessageInformation, []byte{}),
										},
									},
								},
							},
						},
					},
				},
			},
			expectOutput: &wire.SNACMessage{
				Frame: wire.SNACFrame{
					FoodGroup: wire.Chat,
					SubGroup:  wire.ChatChannelMsgToClient,
					RequestID: 1234,
				},
				Body: wire.SNAC_0x0E_0x06_ChatChannelMsgToClient{
					Cookie:  1234,
					Channel: 14,
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							wire.NewTLVBE(wire.ChatTLVSenderInformation,
								newTestSession("user_sending_chat_msg", sessOptCannedSignonTime).TLVUserInfo()),
							wire.NewTLVBE(wire.ChatTLVPublicWhisperFlag, []byte{}),
							wire.NewTLVBE(wire.ChatTLVMessageInformation, []byte{}),
						},
					},
				},
			},
		},
		{
			name: "send chat room message with macOS client 4.0.9 bug containing bad channel ID, expect message to " +
				"client on MIME channel",
			userSession: newTestSession("user_sending_chat_msg", sessOptCannedSignonTime,
				sessOptChatRoomCookie("the-chat-cookie")),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x0E_0x05_ChatChannelMsgToHost{
					Cookie:  1234,
					Channel: math.MaxUint16,
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							{
								Tag:   wire.ChatTLVPublicWhisperFlag,
								Value: []byte{},
							},
							{
								Tag:   wire.ChatTLVMessageInformation,
								Value: []byte{},
							},
						},
					},
				},
			},
			mockParams: mockParams{
				chatMessageRelayerParams: chatMessageRelayerParams{
					chatRelayToAllExceptParams: chatRelayToAllExceptParams{
						{
							screenName: state.NewIdentScreenName("user_sending_chat_msg"),
							cookie:     "the-chat-cookie",
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Chat,
									SubGroup:  wire.ChatChannelMsgToClient,
								},
								Body: wire.SNAC_0x0E_0x06_ChatChannelMsgToClient{
									Cookie:  1234,
									Channel: wire.ICBMChannelMIME,
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLVBE(wire.ChatTLVSenderInformation,
												newTestSession("user_sending_chat_msg", sessOptCannedSignonTime).TLVUserInfo()),
											wire.NewTLVBE(wire.ChatTLVPublicWhisperFlag, []byte{}),
											wire.NewTLVBE(wire.ChatTLVMessageInformation, []byte{}),
										},
									},
								},
							},
						},
					},
				},
			},
			expectOutput: nil,
		},
		{
			name: "send chat room message, don't expect acknowledgement to sender client",
			userSession: newTestSession("user_sending_chat_msg", sessOptCannedSignonTime,
				sessOptChatRoomCookie("the-chat-cookie")),
			inputSNAC: wire.SNACMessage{
				Frame: wire.SNACFrame{
					RequestID: 1234,
				},
				Body: wire.SNAC_0x0E_0x05_ChatChannelMsgToHost{
					Cookie:  1234,
					Channel: 14,
					TLVRestBlock: wire.TLVRestBlock{
						TLVList: wire.TLVList{
							{
								Tag:   wire.ChatTLVPublicWhisperFlag,
								Value: []byte{},
							},
							{
								Tag:   wire.ChatTLVMessageInformation,
								Value: []byte{},
							},
						},
					},
				},
			},
			mockParams: mockParams{
				chatMessageRelayerParams: chatMessageRelayerParams{
					chatRelayToAllExceptParams: chatRelayToAllExceptParams{
						{
							screenName: state.NewIdentScreenName("user_sending_chat_msg"),
							cookie:     "the-chat-cookie",
							message: wire.SNACMessage{
								Frame: wire.SNACFrame{
									FoodGroup: wire.Chat,
									SubGroup:  wire.ChatChannelMsgToClient,
								},
								Body: wire.SNAC_0x0E_0x06_ChatChannelMsgToClient{
									Cookie:  1234,
									Channel: 14,
									TLVRestBlock: wire.TLVRestBlock{
										TLVList: wire.TLVList{
											wire.NewTLVBE(wire.ChatTLVSenderInformation,
												newTestSession("user_sending_chat_msg", sessOptCannedSignonTime).TLVUserInfo()),
											wire.NewTLVBE(wire.ChatTLVPublicWhisperFlag, []byte{}),
											wire.NewTLVBE(wire.ChatTLVMessageInformation, []byte{}),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chatMessageRelayer := newMockChatMessageRelayer(t)
			for _, params := range tc.mockParams.chatRelayToAllExceptParams {
				chatMessageRelayer.EXPECT().
					RelayToAllExcept(mock.Anything, params.cookie, params.screenName, params.message)
			}

			svc := NewChatService(chatMessageRelayer)
			outputSNAC, err := svc.ChannelMsgToHost(context.Background(), tc.userSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(wire.SNAC_0x0E_0x05_ChatChannelMsgToHost))
			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}
