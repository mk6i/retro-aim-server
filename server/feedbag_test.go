package server

import (
	"bytes"
	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestReceiveAndSendFeedbagQuery(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// screenName is the buddy list owner
		screenName string
		// feedbagItems is the list of items in user's buddy list
		feedbagItems []oscar.FeedbagItem
		// lastModified is the time the buddy list was last changed
		lastModified time.Time
		// inputSNAC is the SNAC frame sent from the server to the recipient
		// client
		expectSNACFrame oscar.SnacFrame
		// expectSNACBody is the SNAC payload sent from the server to the
		// recipient client
		expectSNACBody oscar.SNAC_0x13_0x06_FeedbagReply
	}{
		{
			name:         "retrieve empty feedbag",
			screenName:   "sender-screen-name",
			feedbagItems: []oscar.FeedbagItem{},
			lastModified: time.UnixMilli(0),
			expectSNACFrame: oscar.SnacFrame{
				FoodGroup: FEEDBAG,
				SubGroup:  FeedbagReply,
			},
			expectSNACBody: oscar.SNAC_0x13_0x06_FeedbagReply{},
		},
		{
			name:       "retrieve feedbag with items",
			screenName: "sender-screen-name",
			feedbagItems: []oscar.FeedbagItem{
				{
					Name: "buddy_1",
				},
				{
					Name: "buddy_2",
				},
			},
			lastModified: time.UnixMilli(1696472198082),
			expectSNACFrame: oscar.SnacFrame{
				FoodGroup: FEEDBAG,
				SubGroup:  FeedbagReply,
			},
			expectSNACBody: oscar.SNAC_0x13_0x06_FeedbagReply{
				Version: 0,
				Items: []oscar.FeedbagItem{
					{
						Name: "buddy_1",
					},
					{
						Name: "buddy_2",
					},
				},
				LastUpdate: uint32(time.UnixMilli(1696472198082).Unix()),
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
				Retrieve(tc.screenName).
				Return(tc.feedbagItems, nil).
				Maybe()
			fm.EXPECT().
				LastModified(tc.screenName).
				Return(tc.lastModified, nil).
				Maybe()
			//
			// send input SNAC
			//
			var seq uint32
			snac := oscar.SnacFrame{
				FoodGroup: FEEDBAG,
				SubGroup:  FeedbagQuery,
			}
			senderSession := &Session{
				ScreenName: tc.screenName,
			}
			output := &bytes.Buffer{}
			assert.NoError(t, ReceiveAndSendFeedbagQuery(senderSession, fm, snac, output, &seq))
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
			actual := oscar.SNAC_0x13_0x06_FeedbagReply{}
			assert.NoError(t, oscar.Unmarshal(&actual, output))
			assert.Equal(t, tc.expectSNACBody, actual)
		})
	}
}

func TestReceiveAndSendFeedbagQueryIfModified(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// screenName is the buddy list owner
		screenName string
		// feedbagItems is the list of items in user's buddy list
		feedbagItems []oscar.FeedbagItem
		// lastModified is the time the buddy list was last changed
		lastModified time.Time
		// inputSNAC is the SNAC frame sent from the server to the recipient
		// client
		expectSNACFrame oscar.SnacFrame
		// inputSNAC is the SNAC sent by the sender client
		inputSNAC oscar.SNAC_0x13_0x05_FeedbagQueryIfModified
		// expectSNACBody is the SNAC payload sent from the server to the
		// recipient client
		expectSNACBody any
	}{
		{
			name:         "retrieve empty feedbag",
			screenName:   "sender-screen-name",
			feedbagItems: []oscar.FeedbagItem{},
			lastModified: time.UnixMilli(0),
			expectSNACFrame: oscar.SnacFrame{
				FoodGroup: FEEDBAG,
				SubGroup:  FeedbagReply,
			},
			inputSNAC: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
				LastUpdate: uint32(time.UnixMilli(100000).Unix()),
			},
			expectSNACBody: oscar.SNAC_0x13_0x06_FeedbagReply{},
		},
		{
			name:       "retrieve feedbag with items",
			screenName: "sender-screen-name",
			feedbagItems: []oscar.FeedbagItem{
				{
					Name: "buddy_1",
				},
				{
					Name: "buddy_2",
				},
			},
			lastModified: time.UnixMilli(200000),
			expectSNACFrame: oscar.SnacFrame{
				FoodGroup: FEEDBAG,
				SubGroup:  FeedbagReply,
			},
			inputSNAC: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
				LastUpdate: uint32(time.UnixMilli(100000).Unix()),
			},
			expectSNACBody: oscar.SNAC_0x13_0x06_FeedbagReply{
				Version: 0,
				Items: []oscar.FeedbagItem{
					{
						Name: "buddy_1",
					},
					{
						Name: "buddy_2",
					},
				},
				LastUpdate: uint32(time.UnixMilli(200000).Unix()),
			},
		},
		{
			name:       "retrieve not-modified response",
			screenName: "sender-screen-name",
			feedbagItems: []oscar.FeedbagItem{
				{
					Name: "buddy_1",
				},
				{
					Name: "buddy_2",
				},
			},
			lastModified: time.UnixMilli(100000),
			expectSNACFrame: oscar.SnacFrame{
				FoodGroup: FEEDBAG,
				SubGroup:  FeedbagReplyNotModified,
			},
			inputSNAC: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
				LastUpdate: uint32(time.UnixMilli(200000).Unix()),
			},
			expectSNACBody: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
				LastUpdate: uint32(time.UnixMilli(100000).Unix()),
				Count:      2,
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
				Retrieve(tc.screenName).
				Return(tc.feedbagItems, nil).
				Maybe()
			fm.EXPECT().
				LastModified(tc.screenName).
				Return(tc.lastModified, nil).
				Maybe()
			//
			// send input SNAC
			//
			input := &bytes.Buffer{}
			var seq uint32
			assert.NoError(t, oscar.Marshal(tc.inputSNAC, input))
			output := &bytes.Buffer{}
			snac := oscar.SnacFrame{
				FoodGroup: FEEDBAG,
				SubGroup:  FeedbagQuery,
			}
			senderSession := &Session{
				ScreenName: tc.screenName,
			}
			assert.NoError(t, ReceiveAndSendFeedbagQueryIfModified(senderSession, fm, snac, input, output, &seq))
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
			case oscar.SNAC_0x13_0x06_FeedbagReply:
				outputSNAC := oscar.SNAC_0x13_0x06_FeedbagReply{}
				assert.NoError(t, oscar.Unmarshal(&outputSNAC, output))
				assert.Equal(t, v, outputSNAC)
			case oscar.SNAC_0x13_0x05_FeedbagQueryIfModified:
				outputSNAC := oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{}
				assert.NoError(t, oscar.Unmarshal(&outputSNAC, output))
				assert.Equal(t, v, outputSNAC)
			default:
				t.Fatalf("unexpected output SNAC type")
			}
		})
	}
}
