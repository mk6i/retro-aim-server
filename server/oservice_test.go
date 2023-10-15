package server

import (
	"bytes"
	"testing"
	"time"

	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
)

func TestReceiveAndSendServiceRequest(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// cfg is the application config
		cfg Config
		// chatRoom is the chat room the user connects to
		chatRoom *ChatRoom
		// userSession is the session of the user requesting the chat service
		// info
		userSession *Session
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC oscar.SNAC_0x01_0x04_OServiceServiceRequest
		// expectSNACFrame is the SNAC frame sent from the server to the recipient
		// client
		expectSNACFrame oscar.SnacFrame
		// expectSNACBody is the SNAC payload sent from the server to the
		// recipient client
		expectSNACBody any
	}{
		{
			name: "request info for ICBM service, return invalid SNAC err",
			userSession: &Session{
				ScreenName: "user_screen_name",
			},
			inputSNAC: oscar.SNAC_0x01_0x04_OServiceServiceRequest{
				FoodGroup: ICBM,
			},
			expectSNACFrame: oscar.SnacFrame{
				FoodGroup: OSERVICE,
				SubGroup:  OServiceErr,
			},
			expectSNACBody: oscar.SnacOServiceErr{
				Code: ErrorCodeInvalidSnac,
			},
		},
		{
			name: "request info for connecting to chat room, return chat service and chat room metadata",
			cfg: Config{
				OSCARHost: "127.0.0.1",
				ChatPort:  1234,
			},
			chatRoom: &ChatRoom{
				CreateTime:     time.UnixMilli(0),
				DetailLevel:    4,
				Exchange:       8,
				Cookie:         "the-chat-cookie",
				InstanceNumber: 16,
				Name:           "my new chat",
			},
			userSession: &Session{
				ID:         "user-sess-id",
				ScreenName: "user_screen_name",
			},
			inputSNAC: oscar.SNAC_0x01_0x04_OServiceServiceRequest{
				FoodGroup: CHAT,
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						{
							TType: 0x01,
							Val: oscar.SNAC_0x01_0x04_TLVRoomInfo{
								Exchange:       8,
								Cookie:         []byte("the-chat-cookie"),
								InstanceNumber: 16,
							},
						},
					},
				},
			},
			expectSNACFrame: oscar.SnacFrame{
				FoodGroup: OSERVICE,
				SubGroup:  OServiceServiceResponse,
			},
			expectSNACBody: oscar.SNAC_0x01_0x05_OServiceServiceResponse{
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						{
							TType: OserviceTlvTagsReconnectHere,
							Val:   "127.0.0.1:1234",
						},
						{
							TType: OserviceTlvTagsLoginCookie,
							Val: ChatCookie{
								Cookie: []byte("the-chat-cookie"),
								SessID: "user-sess-id",
							},
						},
						{
							TType: OserviceTlvTagsGroupId,
							Val:   CHAT,
						},
						{
							TType: OserviceTlvTagsSslCertname,
							Val:   "",
						},
						{
							TType: OserviceTlvTagsSslState,
							Val:   uint8(0x00),
						},
					},
				},
			},
		},
		{
			name: "request info for connecting to non-existent chat room, return SNAC error",
			cfg: Config{
				OSCARHost: "127.0.0.1",
				ChatPort:  1234,
			},
			chatRoom: nil,
			userSession: &Session{
				ID:         "user-sess-id",
				ScreenName: "user_screen_name",
			},
			inputSNAC: oscar.SNAC_0x01_0x04_OServiceServiceRequest{
				FoodGroup: CHAT,
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						{
							TType: 0x01,
							Val: oscar.SNAC_0x01_0x04_TLVRoomInfo{
								Exchange:       8,
								Cookie:         []byte("the-chat-cookie"),
								InstanceNumber: 16,
							},
						},
					},
				},
			},
			expectSNACFrame: oscar.SnacFrame{
				FoodGroup: OSERVICE,
				SubGroup:  OServiceErr,
			},
			expectSNACBody: oscar.SnacOServiceErr{
				Code: ErrorCodeInvalidSnac,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			//
			// initialize dependencies
			//
			sm := NewMockSessionManager(t)
			cr := NewChatRegistry()
			if tc.chatRoom != nil {
				sm.EXPECT().
					NewSessionWithSN(tc.userSession.ID, tc.userSession.ScreenName).
					Return(&Session{}).
					Maybe()
				tc.chatRoom.SessionManager = sm
				cr.Register(*tc.chatRoom)
			}

			//
			// send input SNAC
			//
			snac := oscar.SnacFrame{
				FoodGroup: OSERVICE,
				SubGroup:  OServiceServiceRequest,
			}
			input := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.inputSNAC, input))
			output := &bytes.Buffer{}
			var seq uint32
			assert.NoError(t, ReceiveAndSendServiceRequest(tc.cfg, cr, tc.userSession, snac, input, output, &seq))

			//
			// verify server response
			//
			flapFrame := oscar.FlapFrame{}
			assert.NoError(t, oscar.Unmarshal(&flapFrame, output))

			snacFrame := oscar.SnacFrame{}
			assert.NoError(t, oscar.Unmarshal(&snacFrame, output))
			assert.Equal(t, tc.expectSNACFrame, snacFrame)

			switch expectSNAC := tc.expectSNACBody.(type) {
			case oscar.SNAC_0x01_0x05_OServiceServiceResponse:
				assert.NoError(t, expectSNAC.SerializeInPlace())
				outputSNAC := oscar.SNAC_0x01_0x05_OServiceServiceResponse{}
				assert.NoError(t, oscar.Unmarshal(&outputSNAC, output))
				assert.Equal(t, expectSNAC, outputSNAC)
			case oscar.SnacOServiceErr:
				outputSNAC := oscar.SnacOServiceErr{}
				assert.NoError(t, oscar.Unmarshal(&outputSNAC, output))
				assert.Equal(t, expectSNAC, outputSNAC)
			default:
				t.Fatalf("unexpected output SNAC type")
			}
			assert.Equalf(t, 0, output.Len(), "the rest of the buffer is unread")
		})
	}
}
