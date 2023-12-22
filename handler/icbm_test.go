package handler

import (
	"testing"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendAndReceiveChannelMsgToHost(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// blockedState is the response to the sender/recipient block check
		blockedState state.BlockedState
		// recipRetrieveErr is the error returned by the recipient session
		// lookup
		recipRetrieveErr error
		senderSession    *state.Session
		recipientSession *state.Session
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC oscar.SNACMessage
		// expectSNACToClient is the SNAC sent from the server to the
		// recipient client
		expectSNACToClient oscar.SNACMessage
		// inputSNAC is the SNAC frame sent from the server to the recipient
		// client
		expectOutput *oscar.SNACMessage
	}{
		{
			name:             "transmit message from sender to recipient, ack message back to sender",
			blockedState:     state.BlockedNo,
			senderSession:    newTestSession("sender-screen-name", sessOptWarning(10)),
			recipientSession: newTestSession("recipient-screen-name", sessOptWarning(20)),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ScreenName: "recipient-screen-name",
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								Tag:   oscar.ICBMTLVTagRequestHostAck,
								Value: []byte{},
							},
						},
					},
				},
			},
			expectSNACToClient: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMChannelMsgToClient,
				},
				Body: oscar.SNAC_0x04_0x07_ICBMChannelMsgToClient{
					TLVUserInfo: oscar.TLVUserInfo{
						ScreenName:   "sender-screen-name",
						WarningLevel: 10,
					},
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								Tag:   oscar.ICBMTLVTagsWantEvents,
								Value: []byte{},
							},
							{
								Tag:   oscar.ICBMTLVTagRequestHostAck,
								Value: []byte{},
							},
						},
					},
				},
			},
			expectOutput: &oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMHostAck,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x04_0x0C_ICBMHostAck{
					ScreenName: "recipient-screen-name",
				},
			},
		},
		{
			name:             "transmit message from sender to recipient, don't ack message back to sender",
			blockedState:     state.BlockedNo,
			senderSession:    newTestSession("sender-screen-name", sessOptWarning(10)),
			recipientSession: newTestSession("recipient-screen-name", sessOptWarning(20)),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ScreenName: "recipient-screen-name",
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{},
					},
				},
			},
			expectSNACToClient: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMChannelMsgToClient,
				},
				Body: oscar.SNAC_0x04_0x07_ICBMChannelMsgToClient{
					TLVUserInfo: oscar.TLVUserInfo{
						ScreenName:   "sender-screen-name",
						WarningLevel: 10,
					},
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								Tag:   oscar.ICBMTLVTagsWantEvents,
								Value: []byte{},
							},
						},
					},
				},
			},
			expectOutput: nil,
		},
		{
			name:             "don't transmit message from sender to recipient because sender has blocked recipient",
			blockedState:     state.BlockedA,
			senderSession:    newTestSession("sender-screen-name", sessOptWarning(10)),
			recipientSession: newTestSession("recipient-screen-name", sessOptWarning(20)),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ScreenName: "recipient-screen-name",
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								Tag:   oscar.ICBMTLVTagRequestHostAck,
								Value: []byte{},
							},
						},
					},
				},
			},
			expectOutput: &oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMErr,
					RequestID: 1234,
				},
				Body: oscar.SNACError{
					Code: oscar.ErrorCodeInLocalPermitDeny,
				},
			},
		},
		{
			name:             "don't transmit message from sender to recipient because recipient has blocked sender",
			blockedState:     state.BlockedB,
			senderSession:    newTestSession("sender-screen-name", sessOptWarning(10)),
			recipientSession: newTestSession("recipient-screen-name", sessOptWarning(20)),
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ScreenName: "recipient-screen-name",
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								Tag:   oscar.ICBMTLVTagRequestHostAck,
								Value: []byte{},
							},
						},
					},
				},
			},
			expectOutput: &oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMErr,
					RequestID: 1234,
				},
				Body: oscar.SNACError{
					Code: oscar.ErrorCodeNotLoggedOn,
				},
			},
		},
		{
			name:             "don't transmit message from sender to recipient because recipient doesn't exist",
			blockedState:     state.BlockedNo,
			senderSession:    newTestSession("sender-screen-name", sessOptWarning(10)),
			recipientSession: nil,
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ScreenName: "recipient-screen-name",
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								Tag:   oscar.ICBMTLVTagRequestHostAck,
								Value: []byte{},
							},
						},
					},
				},
			},
			expectOutput: &oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMErr,
					RequestID: 1234,
				},
				Body: oscar.SNACError{
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
			feedbagManager := newMockFeedbagManager(t)
			feedbagManager.EXPECT().
				BlockedState(tc.senderSession.ScreenName(),
					tc.inputSNAC.Body.(oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost).ScreenName).
				Return(tc.blockedState, nil).
				Maybe()
			messageRelayer := newMockMessageRelayer(t)
			messageRelayer.EXPECT().
				RetrieveByScreenName(tc.inputSNAC.Body.(oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost).ScreenName).
				Return(tc.recipientSession).
				Maybe()
			if tc.recipientSession != nil {
				messageRelayer.EXPECT().
					RelayToScreenName(mock.Anything, tc.recipientSession.ScreenName(), tc.expectSNACToClient).
					Maybe()
			}
			//
			// send input SNAC
			//
			svc := ICBMService{
				messageRelayer: messageRelayer,
				feedbagManager: feedbagManager,
			}
			outputSNAC, err := svc.ChannelMsgToHostHandler(nil, tc.senderSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost))
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
		blockedState state.BlockedState
		// senderScreenName is the screen name of the user sending the event
		senderScreenName string
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC oscar.SNACMessage
		// expectSNACToClient is the SNAC sent from the server to the
		// recipient client
		expectSNACToClient oscar.SNACMessage
	}{
		{
			name:             "transmit message from sender to recipient",
			blockedState:     state.BlockedNo,
			senderScreenName: "sender-screen-name",
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x04_0x14_ICBMClientEvent{
					Cookie:     [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
					ChannelID:  42,
					ScreenName: "recipient-screen-name",
					Event:      12,
				},
			},
			expectSNACToClient: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMClientEvent,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x04_0x14_ICBMClientEvent{
					Cookie:     [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
					ChannelID:  42,
					ScreenName: "sender-screen-name",
					Event:      12,
				},
			},
		},
		{
			name:             "don't transmit message from sender to recipient because sender has blocked recipient",
			blockedState:     state.BlockedA,
			senderScreenName: "sender-screen-name",
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x04_0x14_ICBMClientEvent{
					ScreenName: "recipient-screen-name",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			//
			// initialize dependencies
			//
			feedbagManager := newMockFeedbagManager(t)
			feedbagManager.EXPECT().
				BlockedState(tc.senderScreenName, tc.inputSNAC.Body.(oscar.SNAC_0x04_0x14_ICBMClientEvent).ScreenName).
				Return(tc.blockedState, nil).
				Maybe()
			messageRelayer := newMockMessageRelayer(t)
			if tc.blockedState == state.BlockedNo {
				messageRelayer.EXPECT().
					RelayToScreenName(mock.Anything, tc.inputSNAC.Body.(oscar.SNAC_0x04_0x14_ICBMClientEvent).ScreenName,
						tc.expectSNACToClient)
			}
			//
			// send input SNAC
			//
			senderSession := newTestSession(tc.senderScreenName)
			svc := ICBMService{
				messageRelayer: messageRelayer,
				feedbagManager: feedbagManager,
			}
			assert.NoError(t, svc.ClientEventHandler(nil, senderSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(oscar.SNAC_0x04_0x14_ICBMClientEvent)))
		})
	}
}

func TestSendAndReceiveEvilRequest(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// blockedState is the response to the sender/recipient block check
		blockedState state.BlockedState
		// recipRetrieveErr is the error returned by the recipient session
		// lookup
		recipRetrieveErr error
		// senderScreenName is the session name of the user sending the IM
		senderSession *state.Session
		// recipientScreenName is the screen name of the user receiving the IM
		recipientScreenName string
		// recipientBuddies is a list of the recipient's buddies that get
		// updated warning level
		recipientBuddies []string
		broadcastMessage oscar.SNACMessage
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC oscar.SNACMessage
		// expectSNACToClient is the SNAC sent from the server to the
		// recipient client
		expectSNACToClient oscar.SNACMessage

		expectOutput oscar.SNACMessage
	}{
		{
			name:                "transmit anonymous warning from sender to recipient",
			blockedState:        state.BlockedNo,
			senderSession:       newTestSession("sender-screen-name"),
			recipientScreenName: "recipient-screen-name",
			broadcastMessage: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Buddy,
					SubGroup:  oscar.BuddyArrived,
				},
				Body: oscar.SNAC_0x03_0x0B_BuddyArrived{
					TLVUserInfo: newTestSession("recipient-screen-name", sessOptCannedSignonTime, sessOptWarning(evilDeltaAnon)).TLVUserInfo(),
				},
			},
			recipientBuddies: []string{"buddy1", "buddy2"},
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x04_0x08_ICBMEvilRequest{
					SendAs:     1, // make it anonymous
					ScreenName: "recipient-screen-name",
				},
			},
			expectSNACToClient: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceEvilNotification,
				},
				Body: oscar.SNAC_0x01_0x10_OServiceEvilNotificationAnon{
					NewEvil: evilDeltaAnon,
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMEvilReply,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x04_0x09_ICBMEvilReply{
					EvilDeltaApplied: 30,
					UpdatedEvilValue: 30,
				},
			},
		},
		{
			name:                "transmit non-anonymous warning from sender to recipient",
			blockedState:        state.BlockedNo,
			senderSession:       newTestSession("sender-screen-name"),
			recipientScreenName: "recipient-screen-name",
			recipientBuddies:    []string{"buddy1", "buddy2"},
			broadcastMessage: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Buddy,
					SubGroup:  oscar.BuddyArrived,
				},
				Body: oscar.SNAC_0x03_0x0B_BuddyArrived{
					TLVUserInfo: newTestSession("recipient-screen-name", sessOptCannedSignonTime, sessOptWarning(evilDelta)).TLVUserInfo(),
				},
			},
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x04_0x08_ICBMEvilRequest{
					SendAs:     0, // make it identified
					ScreenName: "recipient-screen-name",
				},
			},
			expectSNACToClient: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.OService,
					SubGroup:  oscar.OServiceEvilNotification,
				},
				Body: oscar.SNAC_0x01_0x10_OServiceEvilNotification{
					NewEvil: evilDelta,
					TLVUserInfo: oscar.TLVUserInfo{
						ScreenName:   "sender-screen-name",
						WarningLevel: 100,
					},
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMEvilReply,
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x04_0x09_ICBMEvilReply{
					EvilDeltaApplied: 100,
					UpdatedEvilValue: 100,
				},
			},
		},
		{
			name:                "don't transmit non-anonymous warning from sender to recipient because sender has blocked recipient",
			blockedState:        state.BlockedA,
			senderSession:       newTestSession("sender-screen-name"),
			recipientScreenName: "recipient-screen-name",
			recipientBuddies:    []string{"buddy1", "buddy2"},
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x04_0x08_ICBMEvilRequest{
					SendAs:     0, // make it identified
					ScreenName: "recipient-screen-name",
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMErr,
					RequestID: 1234,
				},
				Body: oscar.SNACError{
					Code: oscar.ErrorCodeNotLoggedOn,
				},
			},
		},
		{
			name:                "don't transmit non-anonymous warning from sender to recipient because recipient has blocked sender",
			blockedState:        state.BlockedB,
			senderSession:       newTestSession("sender-screen-name"),
			recipientScreenName: "recipient-screen-name",
			recipientBuddies:    []string{"buddy1", "buddy2"},
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x04_0x08_ICBMEvilRequest{
					SendAs:     0, // make it identified
					ScreenName: "recipient-screen-name",
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMErr,
					RequestID: 1234,
				},
				Body: oscar.SNACError{
					Code: oscar.ErrorCodeNotLoggedOn,
				},
			},
		},
		{
			name:                "don't let users block themselves",
			senderSession:       newTestSession("sender-screen-name"),
			recipientScreenName: "sender-screen-name",
			inputSNAC: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					RequestID: 1234,
				},
				Body: oscar.SNAC_0x04_0x08_ICBMEvilRequest{
					SendAs:     0, // make it identified
					ScreenName: "sender-screen-name",
				},
			},
			expectOutput: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMErr,
					RequestID: 1234,
				},
				Body: oscar.SNACError{
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
			feedbagManager := newMockFeedbagManager(t)
			feedbagManager.EXPECT().
				BlockedState(tc.senderSession.ScreenName(), tc.recipientScreenName).
				Return(tc.blockedState, nil).
				Maybe()
			feedbagManager.EXPECT().
				InterestedUsers(tc.recipientScreenName).
				Return(tc.recipientBuddies, nil).
				Maybe()
			recipSess := newTestSession(tc.recipientScreenName, sessOptCannedSignonTime)
			messageRelayer := newMockMessageRelayer(t)
			messageRelayer.EXPECT().
				RetrieveByScreenName(tc.recipientScreenName).
				Return(recipSess).
				Maybe()
			messageRelayer.EXPECT().
				RelayToScreenName(mock.Anything, tc.recipientScreenName, tc.expectSNACToClient).
				Maybe()
			messageRelayer.EXPECT().
				RelayToScreenNames(mock.Anything, tc.recipientBuddies, tc.broadcastMessage).
				Maybe()
			//
			// send input SNAC
			//
			senderSession := newTestSession(tc.senderSession.ScreenName())
			svc := ICBMService{
				messageRelayer: messageRelayer,
				feedbagManager: feedbagManager,
			}
			outputSNAC, err := svc.EvilRequestHandler(nil, senderSession, tc.inputSNAC.Frame,
				tc.inputSNAC.Body.(oscar.SNAC_0x04_0x08_ICBMEvilRequest))
			assert.NoError(t, err)
			assert.Equal(t, tc.expectOutput, outputSNAC)
		})
	}
}
