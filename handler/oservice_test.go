package handler

import (
	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/server"
	"github.com/mkaminski/goaim/state"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestReceiveAndSendServiceRequest(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// config is the application config
		cfg server.Config
		// chatRoom is the chat room the user connects to
		chatRoom *state.ChatRoom
		// userSession is the session of the user requesting the chat service
		// info
		userSession *state.Session
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC oscar.SNAC_0x01_0x04_OServiceServiceRequest
		// expectSNACFrame is the SNAC frame sent from the server to the recipient
		// client
		expectOutput oscar.XMessage
		// expectErr is the expected error returned by the router
		expectErr error
	}{
		{
			name:        "request info for ICBM service, return invalid SNAC err",
			userSession: newTestSession("user_screen_name"),
			inputSNAC: oscar.SNAC_0x01_0x04_OServiceServiceRequest{
				FoodGroup: oscar.ICBM,
			},
			expectErr: server.ErrUnsupportedSubGroup,
		},
		{
			name: "request info for connecting to chat room, return chat service and chat room metadata",
			cfg: server.Config{
				OSCARHost: "127.0.0.1",
				ChatPort:  1234,
			},
			chatRoom: &state.ChatRoom{
				CreateTime:     time.UnixMilli(0),
				DetailLevel:    4,
				Exchange:       8,
				Cookie:         "the-chat-cookie",
				InstanceNumber: 16,
				Name:           "my new chat",
			},
			userSession: newTestSession("user_screen_name", sessOptCannedID),
			inputSNAC: oscar.SNAC_0x01_0x04_OServiceServiceRequest{
				FoodGroup: oscar.CHAT,
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(0x01, oscar.SNAC_0x01_0x04_TLVRoomInfo{
							Exchange:       8,
							Cookie:         []byte("the-chat-cookie"),
							InstanceNumber: 16,
						}),
					},
				},
			},
			expectOutput: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.OSERVICE,
					SubGroup:  oscar.OServiceServiceResponse,
				},
				SnacOut: oscar.SNAC_0x01_0x05_OServiceServiceResponse{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(oscar.OServiceTLVTagsReconnectHere, "127.0.0.1:1234"),
							oscar.NewTLV(oscar.OServiceTLVTagsLoginCookie, server.ChatCookie{
								Cookie: []byte("the-chat-cookie"),
								SessID: "user-sess-id",
							}),
							oscar.NewTLV(oscar.OServiceTLVTagsGroupID, oscar.CHAT),
							oscar.NewTLV(oscar.OServiceTLVTagsSSLCertName, ""),
							oscar.NewTLV(oscar.OServiceTLVTagsSSLState, uint8(0x00)),
						},
					},
				},
			},
		},
		{
			name: "request info for connecting to non-existent chat room, return SNAC error",
			cfg: server.Config{
				OSCARHost: "127.0.0.1",
				ChatPort:  1234,
			},
			chatRoom:    nil,
			userSession: newTestSession("user_screen_name", sessOptCannedID),
			inputSNAC: oscar.SNAC_0x01_0x04_OServiceServiceRequest{
				FoodGroup: oscar.CHAT,
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						oscar.NewTLV(0x01, oscar.SNAC_0x01_0x04_TLVRoomInfo{
							Exchange:       8,
							Cookie:         []byte("the-chat-cookie"),
							InstanceNumber: 16,
						}),
					},
				},
			},
			expectErr: server.ErrUnsupportedSubGroup,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			//
			// initialize dependencies
			//
			sm := newMockChatSessionManager(t)
			cr := state.NewChatRegistry()
			if tc.chatRoom != nil {
				sm.EXPECT().
					NewSessionWithSN(tc.userSession.ID(), tc.userSession.ScreenName()).
					Return(&state.Session{}).
					Maybe()
				cr.Register(*tc.chatRoom, sm)
			}
			//
			// send input SNAC
			//
			svc := OServiceServiceForBOS{
				oServiceService: oServiceService{
					cfg: tc.cfg,
					sm:  sm,
				},
				cr: cr,
			}

			outputSNAC, err := svc.ServiceRequestHandler(nil, tc.userSession, tc.inputSNAC)
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
