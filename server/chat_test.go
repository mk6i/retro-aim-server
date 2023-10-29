package server

import (
	"bytes"
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
)

func TestSendAndReceiveChatChannelMsgToHost(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// userSession is the session of the user sending the chat message
		userSession *Session
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost
		// expectSNACToParticipants is the message the server broadcast to chat
		// room participants (except the sender)
		expectSNACToParticipants XMessage
		expectOutput             *XMessage
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
			expectSNACToParticipants: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.CHAT,
					SubGroup:  oscar.ChatChannelMsgToClient,
				},
				snacOut: oscar.SNAC_0x0E_0x06_ChatChannelMsgToClient{
					Cookie:  1234,
					Channel: 14,
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(oscar.ChatTLVPublicWhisperFlag, []byte{}),
							oscar.NewTLV(oscar.ChatTLVEnableReflectionFlag, []byte{}),
							oscar.NewTLV(oscar.ChatTLVSenderInformation,
								newTestSession(Session{ScreenName: "user_sending_chat_msg"}).GetTLVUserInfo()),
						},
					},
				},
			},
			expectOutput: &XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.CHAT,
					SubGroup:  oscar.ChatChannelMsgToClient,
				},
				snacOut: oscar.SNAC_0x0E_0x06_ChatChannelMsgToClient{
					Cookie:  1234,
					Channel: 14,
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(oscar.ChatTLVPublicWhisperFlag, []byte{}),
							oscar.NewTLV(oscar.ChatTLVEnableReflectionFlag, []byte{}),
							oscar.NewTLV(oscar.ChatTLVSenderInformation, newTestSession(Session{
								ScreenName: "user_sending_chat_msg",
							}).GetTLVUserInfo()),
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
			expectSNACToParticipants: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.CHAT,
					SubGroup:  oscar.ChatChannelMsgToClient,
				},
				snacOut: oscar.SNAC_0x0E_0x06_ChatChannelMsgToClient{
					Cookie:  1234,
					Channel: 14,
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(oscar.ChatTLVPublicWhisperFlag, []byte{}),
							oscar.NewTLV(oscar.ChatTLVSenderInformation,
								newTestSession(Session{
									ScreenName: "user_sending_chat_msg",
								}).GetTLVUserInfo()),
						},
					},
				},
			},
			expectOutput: &XMessage{},
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
			svc := ChatService{}
			outputSNAC, err := svc.ChannelMsgToHostHandler(tc.userSession, crm, tc.inputSNAC)
			assert.NoError(t, err)

			if tc.expectOutput.snacFrame == (oscar.SnacFrame{}) {
				return // handler doesn't return response
			}

			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestChatRouter_RouteChat(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// input is the request payload
		input XMessage
		// output is the response payload
		output *XMessage
		// handlerErr is the mocked handler error response
		handlerErr error
		// expectErr is the expected error returned by the router
		expectErr error
	}{
		{
			name: "receive ChatChannelMsgToHost, return ChatChannelMsgToClient",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.CHAT,
					SubGroup:  oscar.ChatChannelMsgToHost,
				},
				snacOut: oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost{
					Channel: 4,
				},
			},
			output: &XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.CHAT,
					SubGroup:  oscar.ChatChannelMsgToClient,
				},
				snacOut: oscar.SNAC_0x0E_0x06_ChatChannelMsgToClient{
					Channel: 4,
				},
			},
		},
		{
			name: "receive ChatChannelMsgToHost, return no response",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.CHAT,
					SubGroup:  oscar.ChatChannelMsgToHost,
				},
				snacOut: oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost{
					Channel: 4,
				},
			},
			output: nil,
		},
		{
			name: "receive ChatRowListInfo, return ErrUnsupportedSubGroup",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.CHAT,
					SubGroup:  oscar.ChatRowListInfo,
				},
				snacOut: struct{}{},
			},
			output:    nil,
			expectErr: ErrUnsupportedSubGroup,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewMockChatHandler(t)
			svc.EXPECT().
				ChannelMsgToHostHandler(mock.Anything, mock.Anything, tc.input.snacOut).
				Return(tc.output, tc.handlerErr).
				Maybe()

			router := ChatRouter{
				ChatHandler: svc,
			}

			bufIn := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.input.snacOut, bufIn))

			bufOut := &bytes.Buffer{}
			seq := uint32(0)

			err := router.RouteChat(nil, nil, tc.input.snacFrame, bufIn, bufOut, &seq)
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}

			if tc.output == nil {
				// make sure no response was sent
				assert.Empty(t, bufOut.Bytes())
				return
			}

			// verify the FLAP frame
			flap := oscar.FlapFrame{}
			assert.NoError(t, oscar.Unmarshal(&flap, bufOut))

			// make sure the sequence increments
			assert.Equal(t, seq, uint32(1))
			assert.Equal(t, flap.Sequence, uint16(0))

			flapBuf, err := flap.SNACBuffer(bufOut)
			assert.NoError(t, err)

			// verify the SNAC frame
			snacFrame := oscar.SnacFrame{}
			assert.NoError(t, oscar.Unmarshal(&snacFrame, flapBuf))
			assert.Equal(t, tc.output.snacFrame, snacFrame)

			// verify the SNAC message
			snacBuf := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.output.snacOut, snacBuf))
			assert.Equal(t, snacBuf.Bytes(), flapBuf.Bytes())
		})
	}
}
