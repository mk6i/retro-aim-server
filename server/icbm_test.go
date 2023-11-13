package server

import (
	"bytes"
	"testing"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendAndReceiveChannelMsgTohost(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// blockedState is the response to the sender/recipient block check
		blockedState BlockedState
		// recipRetrieveErr is the error returned by the recipient session
		// lookup
		recipRetrieveErr error
		senderSession    *user.Session
		recipientSession *user.Session
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost
		// expectSNACToClient is the SNAC sent from the server to the
		// recipient client
		expectSNACToClient oscar.XMessage
		// inputSNAC is the SNAC frame sent from the server to the recipient
		// client
		expectOutput *oscar.XMessage
	}{
		{
			name:             "transmit message from sender to recipient, ack message back to sender",
			blockedState:     BlockedNo,
			senderSession:    newTestSession("sender-screen-name", sessOptWarning(10)),
			recipientSession: newTestSession("recipient-screen-name", sessOptWarning(20)),
			inputSNAC: oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost{
				ScreenName: "recipient-screen-name",
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						{
							TType: oscar.ICBMTLVTagRequestHostAck,
							Val:   []byte{},
						},
					},
				},
			},
			expectSNACToClient: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMChannelMsgToclient,
				},
				SnacOut: oscar.SNAC_0x04_0x07_ICBMChannelMsgToClient{
					TLVUserInfo: oscar.TLVUserInfo{
						ScreenName:   "sender-screen-name",
						WarningLevel: 10,
					},
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								TType: oscar.ICBMTLVTagsWantEvents,
								Val:   []byte{},
							},
							{
								TType: oscar.ICBMTLVTagRequestHostAck,
								Val:   []byte{},
							},
						},
					},
				},
			},
			expectOutput: &oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMHostAck,
				},
				SnacOut: oscar.SNAC_0x04_0x0C_ICBMHostAck{
					ScreenName: "recipient-screen-name",
				},
			},
		},
		{
			name:             "transmit message from sender to recipient, don't ack message back to sender",
			blockedState:     BlockedNo,
			senderSession:    newTestSession("sender-screen-name", sessOptWarning(10)),
			recipientSession: newTestSession("recipient-screen-name", sessOptWarning(20)),
			inputSNAC: oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost{
				ScreenName: "recipient-screen-name",
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{},
				},
			},
			expectSNACToClient: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMChannelMsgToclient,
				},
				SnacOut: oscar.SNAC_0x04_0x07_ICBMChannelMsgToClient{
					TLVUserInfo: oscar.TLVUserInfo{
						ScreenName:   "sender-screen-name",
						WarningLevel: 10,
					},
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								TType: oscar.ICBMTLVTagsWantEvents,
								Val:   []byte{},
							},
						},
					},
				},
			},
			expectOutput: nil,
		},
		{
			name:             "don't transmit message from sender to recipient because sender has blocked recipient",
			blockedState:     BlockedA,
			senderSession:    newTestSession("sender-screen-name", sessOptWarning(10)),
			recipientSession: newTestSession("recipient-screen-name", sessOptWarning(20)),
			inputSNAC: oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost{
				ScreenName: "recipient-screen-name",
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						{
							TType: oscar.ICBMTLVTagRequestHostAck,
							Val:   []byte{},
						},
					},
				},
			},
			expectOutput: &oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMErr,
				},
				SnacOut: oscar.SnacError{
					Code: oscar.ErrorCodeInLocalPermitDeny,
				},
			},
		},
		{
			name:             "don't transmit message from sender to recipient because recipient has blocked sender",
			blockedState:     BlockedB,
			senderSession:    newTestSession("sender-screen-name", sessOptWarning(10)),
			recipientSession: newTestSession("recipient-screen-name", sessOptWarning(20)),
			inputSNAC: oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost{
				ScreenName: "recipient-screen-name",
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						{
							TType: oscar.ICBMTLVTagRequestHostAck,
							Val:   []byte{},
						},
					},
				},
			},
			expectOutput: &oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMErr,
				},
				SnacOut: oscar.SnacError{
					Code: oscar.ErrorCodeNotLoggedOn,
				},
			},
		},
		{
			name:             "don't transmit message from sender to recipient because recipient doesn't exist",
			blockedState:     BlockedNo,
			recipRetrieveErr: ErrSessNotFound,
			senderSession:    newTestSession("sender-screen-name", sessOptWarning(10)),
			recipientSession: newTestSession("recipient-screen-name", sessOptWarning(20)),
			inputSNAC: oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost{
				ScreenName: "recipient-screen-name",
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						{
							TType: oscar.ICBMTLVTagRequestHostAck,
							Val:   []byte{},
						},
					},
				},
			},
			expectOutput: &oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMErr,
				},
				SnacOut: oscar.SnacError{
					Code: oscar.ErrorCodeNotLoggedOn,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			//
			// initialize dependencies
			//
			fm := NewMockFeedbagManager(t)
			fm.EXPECT().
				Blocked(tc.senderSession.ScreenName(), tc.recipientSession.ScreenName()).
				Return(tc.blockedState, nil).
				Maybe()
			sm := NewMockSessionManager(t)
			sm.EXPECT().
				RetrieveByScreenName(tc.recipientSession.ScreenName()).
				Return(tc.recipientSession, tc.recipRetrieveErr).
				Maybe()
			sm.EXPECT().
				SendToScreenName(mock.Anything, tc.recipientSession.ScreenName(), tc.expectSNACToClient).
				Maybe()
			//
			// send input SNAC
			//
			svc := ICBMService{
				sm: sm,
				fm: fm,
			}
			outputSNAC, err := svc.ChannelMsgToHostHandler(nil, tc.senderSession, tc.inputSNAC)
			assert.NoError(t, err)
			//
			// verify output
			//
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestSendAndReceiveClientEvent(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// blockedState is the response to the sender/recipient block check
		blockedState BlockedState
		// senderScreenName is the screen name of the user sending the event
		senderScreenName string
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC oscar.SNAC_0x04_0x14_ICBMClientEvent
		// expectSNACToClient is the SNAC sent from the server to the
		// recipient client
		expectSNACToClient oscar.XMessage
	}{
		{
			name:             "transmit message from sender to recipient",
			blockedState:     BlockedNo,
			senderScreenName: "sender-screen-name",
			inputSNAC: oscar.SNAC_0x04_0x14_ICBMClientEvent{
				Cookie:     [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
				ChannelID:  42,
				ScreenName: "recipient-screen-name",
				Event:      12,
			},
			expectSNACToClient: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMClientEvent,
				},
				SnacOut: oscar.SNAC_0x04_0x14_ICBMClientEvent{
					Cookie:     [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
					ChannelID:  42,
					ScreenName: "sender-screen-name",
					Event:      12,
				},
			},
		},
		{
			name:             "don't transmit message from sender to recipient because sender has blocked recipient",
			blockedState:     BlockedA,
			senderScreenName: "sender-screen-name",
			inputSNAC: oscar.SNAC_0x04_0x14_ICBMClientEvent{
				ScreenName: "recipient-screen-name",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			//
			// initialize dependencies
			//
			fm := NewMockFeedbagManager(t)
			fm.EXPECT().
				Blocked(tc.senderScreenName, tc.inputSNAC.ScreenName).
				Return(tc.blockedState, nil).
				Maybe()
			sm := NewMockSessionManager(t)
			if tc.blockedState == BlockedNo {
				sm.EXPECT().
					SendToScreenName(mock.Anything, tc.inputSNAC.ScreenName, tc.expectSNACToClient)
			}
			//
			// send input SNAC
			//
			senderSession := newTestSession(tc.senderScreenName)
			svc := ICBMService{
				sm: sm,
				fm: fm,
			}
			assert.NoError(t, svc.ClientEventHandler(nil, senderSession, tc.inputSNAC))
		})
	}
}

func TestSendAndReceiveEvilRequest(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// blockedState is the response to the sender/recipient block check
		blockedState BlockedState
		// recipRetrieveErr is the error returned by the recipient session
		// lookup
		recipRetrieveErr error
		// senderScreenName is the session name of the user sending the IM
		senderSession *user.Session
		// recipientScreenName is the screen name of the user receiving the IM
		recipientScreenName string
		// recipientBuddies is a list of the recipient's buddies that get
		// updated warning level
		recipientBuddies []string
		broadcastMessage oscar.XMessage
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC oscar.SNAC_0x04_0x08_ICBMEvilRequest
		// expectSNACToClient is the SNAC sent from the server to the
		// recipient client
		expectSNACToClient oscar.XMessage

		expectOutput oscar.XMessage
	}{
		{
			name:                "transmit anonymous warning from sender to recipient",
			blockedState:        BlockedNo,
			senderSession:       newTestSession("sender-screen-name"),
			recipientScreenName: "recipient-screen-name",
			broadcastMessage: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.BUDDY,
					SubGroup:  oscar.BuddyArrived,
				},
				SnacOut: oscar.SNAC_0x03_0x0A_BuddyArrived{
					TLVUserInfo: oscar.TLVUserInfo{
						ScreenName:   "recipient-screen-name",
						WarningLevel: evilDeltaAnon,
						TLVBlock: oscar.TLVBlock{
							TLVList: newTestSession("", sessOptCannedSignonTime).UserInfo(),
						},
					},
				},
			},
			recipientBuddies: []string{"buddy1", "buddy2"},
			inputSNAC: oscar.SNAC_0x04_0x08_ICBMEvilRequest{
				SendAs:     1, // make it anonymous
				ScreenName: "recipient-screen-name",
			},
			expectSNACToClient: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.OSERVICE,
					SubGroup:  oscar.OServiceEvilNotification,
				},
				SnacOut: oscar.SNAC_0x01_0x10_OServiceEvilNotificationAnon{
					NewEvil: evilDeltaAnon,
				},
			},
			expectOutput: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMEvilReply,
				},
				SnacOut: oscar.SNAC_0x04_0x09_ICBMEvilReply{
					EvilDeltaApplied: 30,
					UpdatedEvilValue: 30,
				},
			},
		},
		{
			name:                "transmit non-anonymous warning from sender to recipient",
			blockedState:        BlockedNo,
			senderSession:       newTestSession("sender-screen-name"),
			recipientScreenName: "recipient-screen-name",
			recipientBuddies:    []string{"buddy1", "buddy2"},
			broadcastMessage: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.BUDDY,
					SubGroup:  oscar.BuddyArrived,
				},
				SnacOut: oscar.SNAC_0x03_0x0A_BuddyArrived{
					TLVUserInfo: oscar.TLVUserInfo{
						ScreenName:   "recipient-screen-name",
						WarningLevel: evilDelta,
						TLVBlock: oscar.TLVBlock{
							TLVList: newTestSession("", sessOptCannedSignonTime).UserInfo(),
						},
					},
				},
			},
			inputSNAC: oscar.SNAC_0x04_0x08_ICBMEvilRequest{
				SendAs:     0, // make it identified
				ScreenName: "recipient-screen-name",
			},
			expectSNACToClient: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.OSERVICE,
					SubGroup:  oscar.OServiceEvilNotification,
				},
				SnacOut: oscar.SNAC_0x01_0x10_OServiceEvilNotification{
					NewEvil: evilDelta,
					TLVUserInfo: oscar.TLVUserInfo{
						ScreenName:   "sender-screen-name",
						WarningLevel: 100,
					},
				},
			},
			expectOutput: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMEvilReply,
				},
				SnacOut: oscar.SNAC_0x04_0x09_ICBMEvilReply{
					EvilDeltaApplied: 100,
					UpdatedEvilValue: 100,
				},
			},
		},
		{
			name:                "don't transmit non-anonymous warning from sender to recipient because sender has blocked recipient",
			blockedState:        BlockedA,
			senderSession:       newTestSession("sender-screen-name"),
			recipientScreenName: "recipient-screen-name",
			recipientBuddies:    []string{"buddy1", "buddy2"},
			inputSNAC: oscar.SNAC_0x04_0x08_ICBMEvilRequest{
				SendAs:     0, // make it identified
				ScreenName: "recipient-screen-name",
			},
			expectOutput: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMErr,
				},
				SnacOut: oscar.SnacError{
					Code: oscar.ErrorCodeNotLoggedOn,
				},
			},
		},
		{
			name:                "don't transmit non-anonymous warning from sender to recipient because recipient has blocked sender",
			blockedState:        BlockedB,
			senderSession:       newTestSession("sender-screen-name"),
			recipientScreenName: "recipient-screen-name",
			recipientBuddies:    []string{"buddy1", "buddy2"},
			inputSNAC: oscar.SNAC_0x04_0x08_ICBMEvilRequest{
				SendAs:     0, // make it identified
				ScreenName: "recipient-screen-name",
			},
			expectOutput: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMErr,
				},
				SnacOut: oscar.SnacError{
					Code: oscar.ErrorCodeNotLoggedOn,
				},
			},
		},
		{
			name:                "don't let users block themselves",
			senderSession:       newTestSession("sender-screen-name"),
			recipientScreenName: "sender-screen-name",
			inputSNAC: oscar.SNAC_0x04_0x08_ICBMEvilRequest{
				SendAs:     0, // make it identified
				ScreenName: "sender-screen-name",
			},
			expectOutput: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMErr,
				},
				SnacOut: oscar.SnacError{
					Code: oscar.ErrorCodeNotSupportedByHost,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			//
			// initialize dependencies
			//
			fm := NewMockFeedbagManager(t)
			fm.EXPECT().
				Blocked(tc.senderSession.ScreenName(), tc.recipientScreenName).
				Return(tc.blockedState, nil).
				Maybe()
			fm.EXPECT().
				InterestedUsers(tc.recipientScreenName).
				Return(tc.recipientBuddies, nil).
				Maybe()
			recipSess := newTestSession(tc.recipientScreenName, sessOptCannedSignonTime)
			sm := NewMockSessionManager(t)
			sm.EXPECT().
				RetrieveByScreenName(tc.recipientScreenName).
				Return(recipSess, tc.recipRetrieveErr).
				Maybe()
			sm.EXPECT().
				SendToScreenName(mock.Anything, tc.recipientScreenName, tc.expectSNACToClient).
				Maybe()
			sm.EXPECT().
				BroadcastToScreenNames(mock.Anything, tc.recipientBuddies, tc.broadcastMessage).
				Maybe()
			//
			// send input SNAC
			//
			senderSession := newTestSession(tc.senderSession.ScreenName())
			svc := ICBMService{
				sm: sm,
				fm: fm,
			}
			outputSNAC, err := svc.EvilRequestHandler(nil, senderSession, tc.inputSNAC)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}

func TestICBMRouter_RouteICBM(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// input is the request payload
		input oscar.XMessage
		// output is the response payload
		output *oscar.XMessage
		// handlerErr is the mocked handler error response
		handlerErr error
		// expectErr is the expected error returned by the router
		expectErr error
	}{
		{
			name: "receive ICBMAddParameters SNAC, return no response",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMAddParameters,
				},
				SnacOut: oscar.SNAC_0x04_0x02_ICBMAddParameters{
					Channel: 1,
				},
			},
			output: nil,
		},
		{
			name: "receive ICBMParameterQuery, return ICBMParameterReply",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMParameterQuery,
				},
				SnacOut: struct{}{}, // empty SNAC
			},
			output: &oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMParameterReply,
				},
				SnacOut: oscar.SNAC_0x04_0x05_ICBMParameterReply{
					MaxSlots: 100,
				},
			},
		},
		{
			name: "receive ICBMChannelMsgToHost, return ICBMHostAck",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMChannelMsgToHost,
				},
				SnacOut: oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ScreenName: "recipient-screen-name",
				},
			},
			output: &oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMHostAck,
				},
				SnacOut: oscar.SNAC_0x04_0x0C_ICBMHostAck{
					ChannelID: 4,
				},
			},
		},
		{
			name: "receive ICBMChannelMsgToHost, return no reply",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMChannelMsgToHost,
				},
				SnacOut: oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ScreenName: "recipient-screen-name",
				},
			},
			output: nil,
		},
		{
			name: "receive ICBMEvilRequest, return ICBMEvilReply",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMEvilRequest,
				},
				SnacOut: oscar.SNAC_0x04_0x08_ICBMEvilRequest{
					ScreenName: "recipient-screen-name",
				},
			},
			output: &oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMEvilReply,
				},
				SnacOut: oscar.SNAC_0x04_0x09_ICBMEvilReply{
					EvilDeltaApplied: 100,
				},
			},
		},
		{
			name: "receive ICBMClientErr, return no response",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMClientErr,
				},
				SnacOut: oscar.SNAC_0x04_0x0B_ICBMClientErr{
					Code: 4,
				},
			},
			output: nil,
		},
		{
			name: "receive ICBMClientEvent, return no response",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMClientEvent,
				},
				SnacOut: oscar.SNAC_0x04_0x14_ICBMClientEvent{
					ScreenName: "recipient-screen-name",
				},
			},
			output: nil,
		},
		{
			name: "receive ICBMMissedCalls, expect ErrUnsupportedSubGroup",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMMissedCalls,
				},
				SnacOut: struct{}{}, // empty SNAC
			},
			output:    nil,
			expectErr: ErrUnsupportedSubGroup,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewMockICBMHandler(t)
			svc.EXPECT().
				ChannelMsgToHostHandler(mock.Anything, mock.Anything, tc.input.SnacOut).
				Return(tc.output, tc.handlerErr).
				Maybe()
			svc.EXPECT().
				ClientEventHandler(mock.Anything, mock.Anything, tc.input.SnacOut).
				Return(tc.handlerErr).
				Maybe()
			if tc.output != nil {
				svc.EXPECT().
					EvilRequestHandler(mock.Anything, mock.Anything, tc.input.SnacOut).
					Return(*tc.output, tc.handlerErr).
					Maybe()
				svc.EXPECT().
					ParameterQueryHandler(mock.Anything).
					Return(*tc.output).
					Maybe()
			}

			router := ICBMRouter{
				ICBMHandler: svc,
				RouteLogger: RouteLogger{
					Logger: NewLogger(Config{}),
				},
			}

			bufIn := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.input.SnacOut, bufIn))

			bufOut := &bytes.Buffer{}
			seq := uint32(1)

			err := router.RouteICBM(nil, nil, tc.input.SnacFrame, bufIn, bufOut, &seq)
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}

			if tc.output == nil {
				// make sure no response was sent
				assert.Empty(t, bufOut.Bytes())
				return
			}

			// make sure the sequence number was incremented
			assert.Equal(t, uint32(2), seq)

			// verify the FLAP frame
			flap := oscar.FlapFrame{}
			assert.NoError(t, oscar.Unmarshal(&flap, bufOut))
			assert.Equal(t, uint16(1), flap.Sequence)

			// verify the SNAC frame
			snacFrame := oscar.SnacFrame{}
			assert.NoError(t, oscar.Unmarshal(&snacFrame, bufOut))
			assert.Equal(t, tc.output.SnacFrame, snacFrame)

			// verify the SNAC message
			snacBuf := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.output.SnacOut, snacBuf))
			assert.Equal(t, snacBuf.Bytes(), bufOut.Bytes())
		})
	}
}
