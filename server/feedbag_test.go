package server

import (
	"bytes"
	"testing"

	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFeedbagRouter_RouteFeedbag(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// input is the request payload
		input oscar.XMessage
		// output is the response payload
		output oscar.XMessage
		// handlerErr is the mocked handler error response
		handlerErr error
		// expectErr is the expected error returned by the router
		expectErr error
	}{
		{
			name: "receive FeedbagRightsQuery, return FeedbagRightsReply",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagRightsQuery,
				},
				SnacOut: oscar.SNAC_0x13_0x02_FeedbagRightsQuery{
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
			output: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagRightsReply,
				},
				SnacOut: oscar.SNAC_0x13_0x03_FeedbagRightsReply{
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
		},
		{
			name: "receive FeedbagQuery, return FeedbagReply",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagQuery,
				},
				SnacOut: oscar.SNAC_0x13_0x02_FeedbagRightsQuery{
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
			output: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagReply,
				},
				SnacOut: oscar.SNAC_0x13_0x06_FeedbagReply{
					Version: 4,
				},
			},
		},
		{
			name: "receive FeedbagQueryIfModified, return FeedbagRightsReply",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagQueryIfModified,
				},
				SnacOut: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
					LastUpdate: 1234,
				},
			},
			output: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagReply,
				},
				SnacOut: oscar.SNAC_0x13_0x06_FeedbagReply{
					LastUpdate: 1234,
				},
			},
		},
		{
			name: "receive FeedbagUse, return no response",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagUse,
				},
				SnacOut: struct{}{},
			},
			output: oscar.XMessage{},
		},
		{
			name: "receive FeedbagInsertItem, return BuddyArrived and FeedbagStatus",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagInsertItem,
				},
				SnacOut: oscar.SNAC_0x13_0x08_FeedbagInsertItem{
					Items: []oscar.FeedbagItem{
						{
							Name: "my-item",
						},
					},
				},
			},
			output: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagStatus,
				},
				SnacOut: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{1234},
				},
			},
		},
		{
			name: "receive FeedbagUpdateItem, return BuddyArrived and FeedbagStatus",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagUpdateItem,
				},
				SnacOut: oscar.SNAC_0x13_0x09_FeedbagUpdateItem{
					Items: []oscar.FeedbagItem{
						{
							Name: "my-item",
						},
					},
				},
			},
			output: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagStatus,
				},
				SnacOut: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{1234},
				},
			},
		},
		{
			name: "receive FeedbagDeleteItem, return FeedbagStatus",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagDeleteItem,
				},
				SnacOut: oscar.SNAC_0x13_0x0A_FeedbagDeleteItem{
					Items: []oscar.FeedbagItem{
						{
							Name: "my-item",
						},
					},
				},
			},
			output: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagStatus,
				},
				SnacOut: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{1234},
				},
			},
		},
		{
			name: "receive FeedbagStartCluster, return no response",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagStartCluster,
				},
				SnacOut: oscar.SNAC_0x13_0x11_FeedbagStartCluster{
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
			output: oscar.XMessage{},
		},
		{
			name: "receive FeedbagEndCluster, return no response",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagEndCluster,
				},
				SnacOut: struct{}{},
			},
			output: oscar.XMessage{},
		},
		{
			name: "receive FeedbagDeleteUser, return ErrUnsupportedSubGroup",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagDeleteUser,
				},
				SnacOut: struct{}{},
			},
			output:    oscar.XMessage{},
			expectErr: ErrUnsupportedSubGroup,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newMockFeedbagHandler(t)
			svc.EXPECT().
				DeleteItemHandler(mock.Anything, mock.Anything, tc.input.SnacOut).
				Return(tc.output, tc.handlerErr).
				Maybe()
			svc.EXPECT().
				QueryHandler(mock.Anything, mock.Anything).
				Return(tc.output, tc.handlerErr).
				Maybe()
			svc.EXPECT().
				QueryIfModifiedHandler(mock.Anything, mock.Anything, tc.input.SnacOut).
				Return(tc.output, tc.handlerErr).
				Maybe()
			svc.EXPECT().
				RightsQueryHandler(mock.Anything).
				Return(tc.output).
				Maybe()
			svc.EXPECT().
				InsertItemHandler(mock.Anything, mock.Anything, tc.input.SnacOut).
				Return(tc.output, tc.handlerErr).
				Maybe()
			svc.EXPECT().
				UpdateItemHandler(mock.Anything, mock.Anything, tc.input.SnacOut).
				Return(tc.output, tc.handlerErr).
				Maybe()
			svc.EXPECT().
				StartClusterHandler(mock.Anything, tc.input.SnacOut).
				Maybe()

			router := FeedbagRouter{
				FeedbagHandler: svc,
				RouteLogger: RouteLogger{
					Logger: NewLogger(Config{}),
				},
			}

			bufIn := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.input.SnacOut, bufIn))

			bufOut := &bytes.Buffer{}
			seq := uint32(0)

			err := router.RouteFeedbag(nil, nil, tc.input.SnacFrame, bufIn, bufOut, &seq)
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}

			if tc.output.SnacFrame == (oscar.SnacFrame{}) {
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
			assert.Equal(t, tc.output.SnacFrame, snacFrame)

			// verify the SNAC message
			snacBuf := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.output.SnacOut, snacBuf))
			assert.Equal(t, snacBuf.Bytes(), flapBuf.Bytes())
		})
	}
}
