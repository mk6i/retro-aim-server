package server

import (
	"bytes"
	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestLocateRouter_RouteLocate(t *testing.T) {
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
			name: "receive LocateRightsQuery, return LocateRightsReply",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.LOCATE,
					SubGroup:  oscar.LocateRightsQuery,
				},
				SnacOut: struct{}{},
			},
			output: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.LOCATE,
					SubGroup:  oscar.LocateRightsReply,
				},
				SnacOut: oscar.SNAC_0x02_0x03_LocateRightsReply{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(0x01, uint16(1000)),
						},
					},
				},
			},
		},
		{
			name: "receive LocateSetInfo, return no response",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.LOCATE,
					SubGroup:  oscar.LocateSetInfo,
				},
				SnacOut: oscar.SNAC_0x02_0x04_LocateSetInfo{
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
				SnacFrame: oscar.SnacFrame{},
			},
		},
		{
			name: "receive LocateSetDirInfo, return LocateSetDirReply",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.LOCATE,
					SubGroup:  oscar.LocateSetDirInfo,
				},
				SnacOut: oscar.SNAC_0x02_0x09_LocateSetDirInfo{
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
					FoodGroup: oscar.LOCATE,
					SubGroup:  oscar.LocateSetDirReply,
				},
				SnacOut: oscar.SNAC_0x02_0x0A_LocateSetDirReply{
					Result: 1,
				},
			},
		},
		{
			name: "receive LocateGetDirInfo, return no response",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.LOCATE,
					SubGroup:  oscar.LocateGetDirInfo,
				},
				SnacOut: oscar.SNAC_0x02_0x0B_LocateGetDirInfo{
					WatcherScreenNames: "screen-name",
				},
			},
			output: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{},
			},
		},
		{
			name: "receive LocateSetKeywordInfo, return LocateSetKeywordReply",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.LOCATE,
					SubGroup:  oscar.LocateSetKeywordInfo,
				},
				SnacOut: oscar.SNAC_0x02_0x0F_LocateSetKeywordInfo{
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
					FoodGroup: oscar.LOCATE,
					SubGroup:  oscar.LocateSetKeywordReply,
				},
				SnacOut: oscar.SNAC_0x02_0x10_LocateSetKeywordReply{
					Unknown: 1,
				},
			},
		},
		{
			name: "receive LocateUserInfoQuery2, return LocateUserInfoReply",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.LOCATE,
					SubGroup:  oscar.LocateUserInfoQuery2,
				},
				SnacOut: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
					Type2: 1,
				},
			},
			output: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.LOCATE,
					SubGroup:  oscar.LocateUserInfoReply,
				},
				SnacOut: oscar.SNAC_0x02_0x06_LocateUserInfoReply{
					TLVUserInfo: oscar.TLVUserInfo{
						ScreenName: "screen-name",
					},
					LocateInfo: oscar.TLVRestBlock{
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
			name: "receive LocateGetKeywordInfo, expect ErrUnsupportedSubGroup",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.LOCATE,
					SubGroup:  oscar.LocateGetKeywordInfo,
				},
				SnacOut: struct{}{}, // empty SNAC
			},
			output:    oscar.XMessage{}, // empty SNAC
			expectErr: ErrUnsupportedSubGroup,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewMockLocateHandler(t)
			svc.EXPECT().
				RightsQueryHandler(mock.Anything).
				Return(tc.output).
				Maybe()
			svc.EXPECT().
				SetDirInfoHandler(mock.Anything).
				Return(tc.output).
				Maybe()
			svc.EXPECT().
				SetInfoHandler(mock.Anything, mock.Anything, tc.input.SnacOut).
				Return(tc.handlerErr).
				Maybe()
			svc.EXPECT().
				SetKeywordInfoHandler(mock.Anything).
				Return(tc.output).
				Maybe()
			svc.EXPECT().
				UserInfoQuery2Handler(mock.Anything, mock.Anything, tc.input.SnacOut).
				Return(tc.output, tc.handlerErr).
				Maybe()

			router := LocateRouter{
				LocateHandler: svc,
				RouteLogger: RouteLogger{
					Logger: NewLogger(Config{}),
				},
			}

			bufIn := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.input.SnacOut, bufIn))

			bufOut := &bytes.Buffer{}
			seq := uint32(1)

			err := router.RouteLocate(nil, nil, tc.input.SnacFrame, bufIn, bufOut, &seq)
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}

			if tc.output.SnacFrame == (oscar.SnacFrame{}) {
				return // handler doesn't return response
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
