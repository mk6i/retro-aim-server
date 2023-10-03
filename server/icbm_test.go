package server

import (
	"bytes"
	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
	"testing"
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
		// senderScreenName is the screen name of the user sending the IM
		senderScreenName string
		// senderWarning is the warning level of the user sending the IM
		senderWarning uint16
		// recipientScreenName is the screen name of the user receiving the IM
		recipientScreenName string
		// recipientWarning is the warning level of the user receiving the IM
		recipientWarning uint16
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost
		// expectSNACToClient is the SNAC sent from the server to the
		// recipient client
		expectSNACToClient XMessage
		// inputSNAC is the SNAC frame sent from the server to the recipient
		// client
		expectSNACFrame oscar.SnacFrame
		// expectSNACBody is the SNAC payload sent from the server to the
		// recipient client
		expectSNACBody any
	}{
		{
			name:                "transmit message from sender to recipient, ack message back to sender",
			blockedState:        BlockedNo,
			senderScreenName:    "sender-screen-name",
			senderWarning:       10,
			recipientScreenName: "recipient-screen-name",
			recipientWarning:    20,
			inputSNAC: oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost{
				ScreenName: "recipient-screen-name",
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						{
							TType: ICBMTLVTagRequestHostAck,
							Val:   []byte{},
						},
					},
				},
			},
			expectSNACToClient: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: ICBM,
					SubGroup:  ICBMChannelMsgToclient,
				},
				snacOut: oscar.SNAC_0x04_0x07_ICBMChannelMsgToClient{
					TLVUserInfo: oscar.TLVUserInfo{
						ScreenName:   "sender-screen-name",
						WarningLevel: 10,
					},
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								TType: ICBMTLVTagsWantEvents,
								Val:   []byte{},
							},
							{
								TType: ICBMTLVTagRequestHostAck,
								Val:   []byte{},
							},
						},
					},
				},
			},
			expectSNACFrame: oscar.SnacFrame{
				FoodGroup: ICBM,
				SubGroup:  ICBMHostAck,
			},
			expectSNACBody: oscar.SNAC_0x04_0x0C_ICBMHostAck{
				ScreenName: "recipient-screen-name",
			},
		},
		{
			name:                "transmit message from sender to recipient, don't ack message back to sender",
			blockedState:        BlockedNo,
			senderScreenName:    "sender-screen-name",
			senderWarning:       10,
			recipientScreenName: "recipient-screen-name",
			recipientWarning:    20,
			inputSNAC: oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost{
				ScreenName: "recipient-screen-name",
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{},
				},
			},
			expectSNACToClient: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: ICBM,
					SubGroup:  ICBMChannelMsgToclient,
				},
				snacOut: oscar.SNAC_0x04_0x07_ICBMChannelMsgToClient{
					TLVUserInfo: oscar.TLVUserInfo{
						ScreenName:   "sender-screen-name",
						WarningLevel: 10,
					},
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								TType: ICBMTLVTagsWantEvents,
								Val:   []byte{},
							},
						},
					},
				},
			},
		},
		{
			name:                "don't transmit message from sender to recipient because sender has blocked recipient",
			blockedState:        BlockedA,
			senderScreenName:    "sender-screen-name",
			senderWarning:       10,
			recipientScreenName: "recipient-screen-name",
			recipientWarning:    20,
			inputSNAC: oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost{
				ScreenName: "recipient-screen-name",
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						{
							TType: ICBMTLVTagRequestHostAck,
							Val:   []byte{},
						},
					},
				},
			},
			expectSNACFrame: oscar.SnacFrame{
				FoodGroup: ICBM,
				SubGroup:  ICBMErr,
			},
			expectSNACBody: oscar.SnacError{
				Code: ErrorCodeInLocalPermitDeny,
			},
		},
		{
			name:                "don't transmit message from sender to recipient because recipient has blocked sender",
			blockedState:        BlockedB,
			senderScreenName:    "sender-screen-name",
			senderWarning:       10,
			recipientScreenName: "recipient-screen-name",
			recipientWarning:    20,
			inputSNAC: oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost{
				ScreenName: "recipient-screen-name",
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						{
							TType: ICBMTLVTagRequestHostAck,
							Val:   []byte{},
						},
					},
				},
			},
			expectSNACFrame: oscar.SnacFrame{
				FoodGroup: ICBM,
				SubGroup:  ICBMErr,
			},
			expectSNACBody: oscar.SnacError{
				Code: ErrorCodeNotLoggedOn,
			},
		},
		{
			name:                "don't transmit message from sender to recipient because recipient doesn't exist",
			blockedState:        BlockedNo,
			recipRetrieveErr:    errSessNotFound,
			senderScreenName:    "sender-screen-name",
			senderWarning:       10,
			recipientScreenName: "recipient-screen-name",
			recipientWarning:    20,
			inputSNAC: oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost{
				ScreenName: "recipient-screen-name",
				TLVRestBlock: oscar.TLVRestBlock{
					TLVList: oscar.TLVList{
						{
							TType: ICBMTLVTagRequestHostAck,
							Val:   []byte{},
						},
					},
				},
			},
			expectSNACFrame: oscar.SnacFrame{
				FoodGroup: ICBM,
				SubGroup:  ICBMErr,
			},
			expectSNACBody: oscar.SnacError{
				Code: ErrorCodeNotLoggedOn,
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
				Blocked(tc.senderScreenName, tc.recipientScreenName).
				Return(tc.blockedState, nil).
				Maybe()
			sm := NewMockSessionManager(t)
			sm.EXPECT().
				RetrieveByScreenName(tc.recipientScreenName).
				Return(&Session{
					ScreenName: tc.recipientScreenName,
					Warning:    tc.recipientWarning,
				}, tc.recipRetrieveErr).
				Maybe()
			sm.EXPECT().
				SendToScreenName(tc.recipientScreenName, tc.expectSNACToClient).
				Maybe()
			//
			// send input SNAC
			//
			input := &bytes.Buffer{}
			var seq uint32
			assert.NoError(t, oscar.Marshal(tc.inputSNAC, input))
			output := &bytes.Buffer{}
			snac := oscar.SnacFrame{
				FoodGroup: ICBM,
				SubGroup:  ICBMChannelMsgTohost,
			}
			senderSession := &Session{
				ScreenName: tc.senderScreenName,
				Warning:    tc.senderWarning,
			}
			assert.NoError(t, SendAndReceiveChannelMsgTohost(sm, fm, senderSession, snac, input, output, &seq))
			//
			// verify output
			//
			if tc.expectSNACFrame.FoodGroup == 0 {
				// no ack was sent
				return
			}
			flap := oscar.FlapFrame{}
			assert.NoError(t, oscar.Unmarshal(&flap, output))
			SnacFrame := oscar.SnacFrame{}
			assert.NoError(t, oscar.Unmarshal(&SnacFrame, output))
			assert.Equal(t, tc.expectSNACFrame, SnacFrame)
			//
			// verify output SNAC body
			//
			switch v := tc.expectSNACBody.(type) {
			case oscar.SNAC_0x04_0x0C_ICBMHostAck:
				outputSNAC := oscar.SNAC_0x04_0x0C_ICBMHostAck{}
				assert.NoError(t, oscar.Unmarshal(&outputSNAC, output))
				assert.Equal(t, v, outputSNAC)
			case oscar.SnacError:
				outputSNAC := oscar.SnacError{}
				assert.NoError(t, oscar.Unmarshal(&outputSNAC, output))
				assert.Equal(t, v, outputSNAC)
			default:
				t.Fatalf("unexpected output SNAC type")
			}
		})
	}
}
