package server

import (
	"bytes"
	"github.com/stretchr/testify/mock"
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
		expectOutput XMessage
		// expectErr is the expected error returned by the router
		expectErr error
	}{
		{
			name: "request info for ICBM service, return invalid SNAC err",
			userSession: &Session{
				ScreenName: "user_screen_name",
			},
			inputSNAC: oscar.SNAC_0x01_0x04_OServiceServiceRequest{
				FoodGroup: ICBM,
			},
			expectErr: ErrUnsupportedSubGroup,
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
			expectOutput: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: OSERVICE,
					SubGroup:  oscar.OServiceServiceResponse,
				},
				snacOut: oscar.SNAC_0x01_0x05_OServiceServiceResponse{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								TType: oscar.OServiceTLVTagsReconnectHere,
								Val:   "127.0.0.1:1234",
							},
							{
								TType: oscar.OServiceTLVTagsLoginCookie,
								Val: ChatCookie{
									Cookie: []byte("the-chat-cookie"),
									SessID: "user-sess-id",
								},
							},
							{
								TType: oscar.OServiceTLVTagsGroupID,
								Val:   CHAT,
							},
							{
								TType: oscar.OServiceTLVTagsSSLCertName,
								Val:   "",
							},
							{
								TType: oscar.OServiceTLVTagsSSLState,
								Val:   uint8(0x00),
							},
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
			expectErr: ErrUnsupportedSubGroup,
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
			assert.NoError(t, tc.inputSNAC.SerializeInPlace())
			svc := OServiceService{}
			outputSNAC, err := svc.ServiceRequestHandler(tc.cfg, cr, tc.userSession, tc.inputSNAC)
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}
			//
			// verify output
			//
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestOServiceRouter_RouteOService(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// input is the request payload
		input XMessage
		// output is the response payload
		output XMessage
		// handlerErr is the mocked handler error response
		handlerErr error
		// expectErr is the expected error returned by the router
		expectErr error
	}{
		{
			name: "receive OServiceClientOnline, return no response",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: OSERVICE,
					SubGroup:  oscar.OServiceClientOnline,
				},
				snacOut: oscar.SNAC_0x01_0x02_OServiceClientOnline{
					GroupVersions: []struct {
						FoodGroup   uint16
						Version     uint16
						ToolID      uint16
						ToolVersion uint16
					}{
						{
							FoodGroup: 10,
						},
					},
				},
			},
			output: XMessage{},
		},
		{
			name: "receive OServiceServiceRequest, return OServiceServiceResponse",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: OSERVICE,
					SubGroup:  oscar.OServiceServiceRequest,
				},
				snacOut: oscar.SNAC_0x01_0x04_OServiceServiceRequest{
					FoodGroup: 10,
				},
			},
			output: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: OSERVICE,
					SubGroup:  oscar.OServiceServiceResponse,
				},
				snacOut: oscar.SNAC_0x01_0x05_OServiceServiceResponse{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								TType: 0x01,
								Val:   uint16(1000),
							},
						},
					},
				},
			},
		},
		{
			name: "receive OServiceRateParamsQuery, return OServiceRateParamsReply",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: OSERVICE,
					SubGroup:  oscar.OServiceRateParamsQuery,
				},
				snacOut: struct{}{},
			},
			output: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: OSERVICE,
					SubGroup:  oscar.OServiceRateParamsReply,
				},
				snacOut: oscar.SNAC_0x01_0x07_OServiceRateParamsReply{
					RateGroups: []struct {
						ID    uint16
						Pairs []struct {
							FoodGroup uint16
							SubGroup  uint16
						} `count_prefix:"uint16"`
					}{
						{
							ID: 1,
						},
					},
				},
			},
		},
		{
			name: "receive OServiceRateParamsSubAdd, return no response",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: OSERVICE,
					SubGroup:  oscar.OServiceRateParamsSubAdd,
				},
				snacOut: oscar.SNAC_0x01_0x08_OServiceRateParamsSubAdd{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								TType: 0x01,
								Val:   []byte{1, 2, 3, 4},
							},
						},
					},
				},
			},
			output: XMessage{},
		},
		{
			name: "receive OServiceUserInfoQuery, return OServiceUserInfoUpdate",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: OSERVICE,
					SubGroup:  oscar.OServiceUserInfoQuery,
				},
				snacOut: struct{}{},
			},
			output: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: OSERVICE,
					SubGroup:  oscar.OServiceUserInfoUpdate,
				},
				snacOut: oscar.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
					TLVUserInfo: oscar.TLVUserInfo{
						ScreenName: "screen-name",
					},
				},
			},
		},
		{
			name: "receive OServiceIdleNotification, return no response",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: OSERVICE,
					SubGroup:  oscar.OServiceIdleNotification,
				},
				snacOut: oscar.SNAC_0x01_0x11_OServiceIdleNotification{
					IdleTime: 10,
				},
			},
			output: XMessage{},
		},
		{
			name: "receive OServiceClientVersions, return OServiceHostVersions",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: OSERVICE,
					SubGroup:  oscar.OServiceClientVersions,
				},
				snacOut: oscar.SNAC_0x01_0x17_OServiceClientVersions{
					Versions: []uint16{
						10,
					},
				},
			},
			output: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: OSERVICE,
					SubGroup:  oscar.OServiceHostVersions,
				},
				snacOut: oscar.SNAC_0x01_0x18_OServiceHostVersions{
					Versions: []uint16{
						10,
					},
				},
			},
		},
		{
			name: "receive OServiceSetUserInfoFields, return OServiceUserInfoUpdate",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: OSERVICE,
					SubGroup:  oscar.OServiceSetUserInfoFields,
				},
				snacOut: oscar.SNAC_0x01_0x1E_OServiceSetUserInfoFields{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								TType: 0x01,
								Val:   []byte{1, 2, 3, 4},
							},
						},
					},
				},
			},
			output: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: OSERVICE,
					SubGroup:  oscar.OServiceUserInfoUpdate,
				},
				snacOut: oscar.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
					TLVUserInfo: oscar.TLVUserInfo{
						ScreenName: "screen-name",
					},
				},
			},
		},
		{
			name: "receive OServicePauseReq, expect ErrUnsupportedSubGroup",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: OSERVICE,
					SubGroup:  oscar.OServicePauseReq,
				},
				snacOut: struct{}{}, // empty SNAC
			},
			output:    XMessage{}, // empty SNAC
			expectErr: ErrUnsupportedSubGroup,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewMockOServiceHandler(t)
			svc.EXPECT().
				ServiceRequestHandler(mock.Anything, mock.Anything, mock.Anything, tc.input.snacOut).
				Return(tc.output, tc.handlerErr).
				Maybe()
			svc.EXPECT().
				RateParamsQueryHandler().
				Return(tc.output).
				Maybe()
			svc.EXPECT().
				UserInfoQueryHandler(mock.Anything).
				Return(tc.output).
				Maybe()
			svc.EXPECT().
				ClientVersionsHandler(tc.input.snacOut).
				Return(tc.output).
				Maybe()
			svc.EXPECT().
				SetUserInfoFieldsHandler(mock.Anything, mock.Anything, mock.Anything, tc.input.snacOut).
				Return(tc.output, tc.handlerErr).
				Maybe()
			svc.EXPECT().
				ClientOnlineHandler(tc.input.snacOut, mock.Anything, mock.Anything, mock.Anything).
				Return(tc.handlerErr).
				Maybe()
			svc.EXPECT().
				RateParamsSubAddHandler(tc.input.snacOut).
				Maybe()
			svc.EXPECT().
				IdleNotificationHandler(mock.Anything, mock.Anything, mock.Anything, tc.input.snacOut).
				Return(tc.handlerErr).
				Maybe()

			router := OServiceRouter{
				OServiceHandler: svc,
			}

			bufIn := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.input.snacOut, bufIn))

			bufOut := &bytes.Buffer{}
			seq := uint32(1)

			err := router.RouteOService(Config{}, nil, nil, nil, nil, nil, tc.input.snacFrame, bufIn, bufOut, &seq)
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}

			if tc.output == (XMessage{}) {
				// make sure no response was sent
				assert.Empty(t, bufOut.Bytes())
				return
			}

			// verify the FLAP frame
			flap := oscar.FlapFrame{}
			assert.NoError(t, oscar.Unmarshal(&flap, bufOut))

			// make sure the sequence number was incremented
			assert.Equal(t, uint32(2), seq)

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
