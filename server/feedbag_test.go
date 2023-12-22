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
		input oscar.SNACMessage
		// output is the response payload
		output oscar.SNACMessage
		// handlerErr is the mocked handler error response
		handlerErr error
		// expectErr is the expected error returned by the router
		expectErr error
	}{
		{
			name: "receive FeedbagRightsQuery, return FeedbagRightsReply",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagRightsQuery,
				},
				Body: oscar.SNAC_0x13_0x02_FeedbagRightsQuery{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								Tag:   0x01,
								Value: []byte{1, 2, 3, 4},
							},
						},
					},
				},
			},
			output: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagRightsReply,
				},
				Body: oscar.SNAC_0x13_0x03_FeedbagRightsReply{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								Tag:   0x01,
								Value: []byte{1, 2, 3, 4},
							},
						},
					},
				},
			},
		},
		{
			name: "receive FeedbagQuery, return FeedbagReply",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagQuery,
				},
				Body: oscar.SNAC_0x13_0x02_FeedbagRightsQuery{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								Tag:   0x01,
								Value: []byte{1, 2, 3, 4},
							},
						},
					},
				},
			},
			output: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagReply,
				},
				Body: oscar.SNAC_0x13_0x06_FeedbagReply{
					Version: 4,
				},
			},
		},
		{
			name: "receive FeedbagQueryIfModified, return FeedbagRightsReply",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagQueryIfModified,
				},
				Body: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
					LastUpdate: 1234,
				},
			},
			output: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagReply,
				},
				Body: oscar.SNAC_0x13_0x06_FeedbagReply{
					LastUpdate: 1234,
				},
			},
		},
		{
			name: "receive FeedbagUse, return no response",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagUse,
				},
				Body: struct{}{},
			},
			output: oscar.SNACMessage{},
		},
		{
			name: "receive FeedbagInsertItem, return FeedbagStatus",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagInsertItem,
				},
				Body: oscar.SNAC_0x13_0x08_FeedbagInsertItem{
					Items: []oscar.FeedbagItem{
						{
							Name: "my-item",
						},
					},
				},
			},
			output: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagStatus,
				},
				Body: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{1234},
				},
			},
		},
		{
			name: "receive FeedbagUpdateItem, return FeedbagStatus",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagUpdateItem,
				},
				Body: oscar.SNAC_0x13_0x09_FeedbagUpdateItem{
					Items: []oscar.FeedbagItem{
						{
							Name: "my-item",
						},
					},
				},
			},
			output: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagStatus,
				},
				Body: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{1234},
				},
			},
		},
		{
			name: "receive FeedbagDeleteItem, return FeedbagStatus",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagDeleteItem,
				},
				Body: oscar.SNAC_0x13_0x0A_FeedbagDeleteItem{
					Items: []oscar.FeedbagItem{
						{
							Name: "my-item",
						},
					},
				},
			},
			output: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagStatus,
				},
				Body: oscar.SNAC_0x13_0x0E_FeedbagStatus{
					Results: []uint16{1234},
				},
			},
		},
		{
			name: "receive FeedbagStartCluster, return no response",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagStartCluster,
				},
				Body: oscar.SNAC_0x13_0x11_FeedbagStartCluster{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							{
								Tag:   0x01,
								Value: []byte{1, 2, 3, 4},
							},
						},
					},
				},
			},
			output: oscar.SNACMessage{},
		},
		{
			name: "receive FeedbagEndCluster, return no response",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagEndCluster,
				},
				Body: struct{}{},
			},
			output: oscar.SNACMessage{},
		},
		{
			name: "receive FeedbagDeleteUser, return ErrUnsupportedSubGroup",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagDeleteUser,
				},
				Body: struct{}{},
			},
			output:    oscar.SNACMessage{},
			expectErr: ErrUnsupportedSubGroup,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newMockFeedbagHandler(t)
			svc.EXPECT().
				DeleteItemHandler(mock.Anything, mock.Anything, tc.input.Frame, tc.input.Body).
				Return(tc.output, tc.handlerErr).
				Maybe()
			svc.EXPECT().
				QueryHandler(mock.Anything, mock.Anything, tc.input.Frame).
				Return(tc.output, tc.handlerErr).
				Maybe()
			svc.EXPECT().
				QueryIfModifiedHandler(mock.Anything, mock.Anything, tc.input.Frame, tc.input.Body).
				Return(tc.output, tc.handlerErr).
				Maybe()
			svc.EXPECT().
				RightsQueryHandler(mock.Anything, tc.input.Frame).
				Return(tc.output).
				Maybe()
			svc.EXPECT().
				InsertItemHandler(mock.Anything, mock.Anything, tc.input.Frame, tc.input.Body).
				Return(tc.output, tc.handlerErr).
				Maybe()
			svc.EXPECT().
				UpdateItemHandler(mock.Anything, mock.Anything, tc.input.Frame, tc.input.Body).
				Return(tc.output, tc.handlerErr).
				Maybe()
			svc.EXPECT().
				StartClusterHandler(mock.Anything, tc.input.Frame, tc.input.Body).
				Maybe()

			router := NewFeedbagRouter(NewLogger(Config{}), svc)

			bufIn := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.input.Body, bufIn))

			bufOut := &bytes.Buffer{}
			seq := uint32(0)

			err := router.RouteFeedbag(nil, nil, tc.input.Frame, bufIn, bufOut, &seq)
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}

			if tc.output.Frame == (oscar.SNACFrame{}) {
				return
			}

			// verify the FLAP frame
			flap := oscar.FLAPFrame{}
			assert.NoError(t, oscar.Unmarshal(&flap, bufOut))

			// make sure the sequence increments
			assert.Equal(t, seq, uint32(1))
			assert.Equal(t, flap.Sequence, uint16(0))

			flapBuf, err := flap.SNACBuffer(bufOut)
			assert.NoError(t, err)

			// verify the SNAC frame
			snacFrame := oscar.SNACFrame{}
			assert.NoError(t, oscar.Unmarshal(&snacFrame, flapBuf))
			assert.Equal(t, tc.output.Frame, snacFrame)

			// verify the SNAC message
			snacBuf := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.output.Body, snacBuf))
			assert.Equal(t, snacBuf.Bytes(), flapBuf.Bytes())
		})
	}
}
