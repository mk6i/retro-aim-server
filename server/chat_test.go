package server

import (
	"bytes"
	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSendAndReceiveChatChannelMsgTohost(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user sending the chat message
		userSession *Session
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost
		// expectSNACFrame is the SNAC frame sent from the server to the
		// recipient client
		expectSNACFrame oscar.SnacFrame
		// expectSNACBody is the SNAC payload sent from the server to the
		// recipient client
		expectSNACBody any
		// expectSNACToParticipants is the message the server broadcast to chat
		// room participants (except the sender)
		expectSNACToParticipants XMessage
	}{
		{
			name: "send chat room message, expect acknowledgement to sender client",
			userSession: newTestSession(Session{
				ScreenName: "user_sending_chat_msg",
			}),
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
			expectSNACFrame: oscar.SnacFrame{
				FoodGroup: CHAT,
				SubGroup:  ChatChannelMsgToclient,
			},
			expectSNACBody: oscar.SNAC_0x0E_0x06_ChatChannelMsgToClient{
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
						{
							TType: oscar.ChatTLVSenderInformation,
							Val: newTestSession(Session{
								ScreenName: "user_sending_chat_msg",
							}).GetTLVUserInfo(),
						},
					},
				},
			},
			expectSNACToParticipants: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: CHAT,
					SubGroup:  ChatChannelMsgToclient,
				},
				snacOut: oscar.SNAC_0x0E_0x06_ChatChannelMsgToClient{
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
							{
								TType: oscar.ChatTLVSenderInformation,
								Val: newTestSession(Session{
									ScreenName: "user_sending_chat_msg",
								}).GetTLVUserInfo(),
							},
						},
					},
				},
			},
		},
		{
			name: "send chat room message, don't expect acknowledgement to sender client",
			userSession: newTestSession(Session{
				ScreenName: "user_sending_chat_msg",
			}),
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
			expectSNACBody: nil,
			expectSNACToParticipants: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: CHAT,
					SubGroup:  ChatChannelMsgToclient,
				},
				snacOut: oscar.SNAC_0x0E_0x06_ChatChannelMsgToClient{
					Cookie:  1234,
					Channel: 14,
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								TType: oscar.ChatTLVPublicWhisperFlag,
								Val:   []byte{},
							},
							{
								TType: oscar.ChatTLVSenderInformation,
								Val: newTestSession(Session{
									ScreenName: "user_sending_chat_msg",
								}).GetTLVUserInfo(),
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			//
			// initialize dependencies
			//
			crm := NewMockSessionManager(t)
			crm.EXPECT().
				BroadcastExcept(tc.userSession, tc.expectSNACToParticipants)
			//
			// send input SNAC
			//
			input := &bytes.Buffer{}
			var seq uint32
			assert.NoError(t, oscar.Marshal(tc.inputSNAC, input))
			output := &bytes.Buffer{}
			snac := oscar.SnacFrame{
				FoodGroup: CHAT,
				SubGroup:  ChatChannelMsgTohost,
			}
			assert.NoError(t, SendAndReceiveChatChannelMsgTohost(tc.userSession, crm, snac, input, output, &seq))
			if tc.expectSNACBody != nil {
				//
				// verify output
				//
				flap := oscar.FlapFrame{}
				assert.NoError(t, oscar.Unmarshal(&flap, output))
				snacFrame := oscar.SnacFrame{}
				assert.NoError(t, oscar.Unmarshal(&snacFrame, output))
				assert.Equal(t, tc.expectSNACFrame, snacFrame)
				//
				// verify output SNAC body
				//
				switch v := tc.expectSNACBody.(type) {
				case oscar.SNAC_0x0E_0x06_ChatChannelMsgToClient:
					assert.NoError(t, v.SerializeInPlace())
					outputSNAC := oscar.SNAC_0x0E_0x06_ChatChannelMsgToClient{}
					assert.NoError(t, oscar.Unmarshal(&outputSNAC, output))
					assert.Equal(t, v, outputSNAC)
				default:
					t.Fatalf("unexpected output SNAC type")
				}
			}
			assert.Equalf(t, 0, output.Len(), "the rest of the buffer is unread")
		})
	}
}
