package server

import (
	"bytes"
	"testing"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/oscar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLocateRouter_RouteLocate(t *testing.T) {
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
			name: "receive LocateRightsQuery, return LocateRightsReply",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Locate,
					SubGroup:  oscar.LocateRightsQuery,
				},
				Body: struct{}{},
			},
			output: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Locate,
					SubGroup:  oscar.LocateRightsReply,
				},
				Body: oscar.SNAC_0x02_0x03_LocateRightsReply{
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
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Locate,
					SubGroup:  oscar.LocateSetInfo,
				},
				Body: oscar.SNAC_0x02_0x04_LocateSetInfo{
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
				Frame: oscar.SNACFrame{},
			},
		},
		{
			name: "receive LocateSetDirInfo, return LocateSetDirReply",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Locate,
					SubGroup:  oscar.LocateSetDirInfo,
				},
				Body: oscar.SNAC_0x02_0x09_LocateSetDirInfo{
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
					FoodGroup: oscar.Locate,
					SubGroup:  oscar.LocateSetDirReply,
				},
				Body: oscar.SNAC_0x02_0x0A_LocateSetDirReply{
					Result: 1,
				},
			},
		},
		{
			name: "receive LocateGetDirInfo, return no response",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Locate,
					SubGroup:  oscar.LocateGetDirInfo,
				},
				Body: oscar.SNAC_0x02_0x0B_LocateGetDirInfo{
					WatcherScreenNames: "screen-name",
				},
			},
			output: oscar.SNACMessage{
				Frame: oscar.SNACFrame{},
			},
		},
		{
			name: "receive LocateSetKeywordInfo, return LocateSetKeywordReply",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Locate,
					SubGroup:  oscar.LocateSetKeywordInfo,
				},
				Body: oscar.SNAC_0x02_0x0F_LocateSetKeywordInfo{
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
					FoodGroup: oscar.Locate,
					SubGroup:  oscar.LocateSetKeywordReply,
				},
				Body: oscar.SNAC_0x02_0x10_LocateSetKeywordReply{
					Unknown: 1,
				},
			},
		},
		{
			name: "receive LocateUserInfoQuery2, return LocateUserInfoReply",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Locate,
					SubGroup:  oscar.LocateUserInfoQuery2,
				},
				Body: oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{
					Type2: 1,
				},
			},
			output: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Locate,
					SubGroup:  oscar.LocateUserInfoReply,
				},
				Body: oscar.SNAC_0x02_0x06_LocateUserInfoReply{
					TLVUserInfo: oscar.TLVUserInfo{
						ScreenName: "screen-name",
					},
					LocateInfo: oscar.TLVRestBlock{
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
			name: "receive LocateGetKeywordInfo, expect ErrUnsupportedSubGroup",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Locate,
					SubGroup:  oscar.LocateGetKeywordInfo,
				},
				Body: struct{}{}, // empty SNAC
			},
			output:    oscar.SNACMessage{}, // empty SNAC
			expectErr: ErrUnsupportedSubGroup,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newMockLocateHandler(t)
			svc.EXPECT().
				RightsQueryHandler(mock.Anything, tc.input.Frame).
				Return(tc.output).
				Maybe()
			svc.EXPECT().
				SetDirInfoHandler(mock.Anything, tc.input.Frame).
				Return(tc.output).
				Maybe()
			svc.EXPECT().
				SetInfoHandler(mock.Anything, mock.Anything, tc.input.Body).
				Return(tc.handlerErr).
				Maybe()
			svc.EXPECT().
				SetKeywordInfoHandler(mock.Anything, tc.input.Frame).
				Return(tc.output).
				Maybe()
			svc.EXPECT().
				UserInfoQuery2Handler(mock.Anything, mock.Anything, tc.input.Frame, tc.input.Body).
				Return(tc.output, tc.handlerErr).
				Maybe()

			router := NewLocateRouter(svc, NewLogger(config.Config{}))

			bufIn := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.input.Body, bufIn))

			bufOut := &bytes.Buffer{}
			seq := uint32(1)

			err := router.Route(nil, nil, tc.input.Frame, bufIn, bufOut, &seq)
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}

			if tc.output.Frame == (oscar.SNACFrame{}) {
				return // handler doesn't return response
			}

			// make sure the sequence number was incremented
			assert.Equal(t, uint32(2), seq)

			// verify the FLAP frame
			flap := oscar.FLAPFrame{}
			assert.NoError(t, oscar.Unmarshal(&flap, bufOut))
			assert.Equal(t, uint16(1), flap.Sequence)

			// verify the SNAC frame
			snacFrame := oscar.SNACFrame{}
			assert.NoError(t, oscar.Unmarshal(&snacFrame, bufOut))
			assert.Equal(t, tc.output.Frame, snacFrame)

			// verify the SNAC message
			snacBuf := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.output.Body, snacBuf))
			assert.Equal(t, snacBuf.Bytes(), bufOut.Bytes())
		})
	}
}
