package server

import (
	"bytes"
	"testing"

	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBuddyRouter_RouteBuddy(t *testing.T) {
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
			name: "receive BuddyRightsQuery, return BuddyRightsReply",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Buddy,
					SubGroup:  oscar.BuddyRightsQuery,
				},
				Body: oscar.SNAC_0x03_0x02_BuddyRightsQuery{
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
			output: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Buddy,
					SubGroup:  oscar.BuddyRightsReply,
				},
				Body: oscar.SNAC_0x03_0x03_BuddyRightsReply{
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
			name: "receive ErrorCodeReplyTooBig, expect ErrUnsupportedSubGroup",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Buddy,
					SubGroup:  oscar.ErrorCodeReplyTooBig,
				},
				Body: struct{}{}, // empty SNAC
			},
			output:    oscar.SNACMessage{},
			expectErr: ErrUnsupportedSubGroup,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newMockBuddyHandler(t)
			svc.EXPECT().
				RightsQueryHandler(mock.Anything).
				Return(tc.output).
				Maybe()

			router := BuddyRouter{
				BuddyHandler: svc,
				RouteLogger: RouteLogger{
					Logger: NewLogger(Config{}),
				},
			}

			bufIn := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.input.Body, bufIn))

			bufOut := &bytes.Buffer{}
			seq := uint32(1)

			err := router.RouteBuddy(nil, tc.input.Frame, bufIn, bufOut, &seq)
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}

			if tc.output == (oscar.SNACMessage{}) {
				// make sure no response was sent
				assert.Empty(t, bufOut.Bytes())
				return
			}

			// verify the FLAP frame
			flap := oscar.FLAPFrame{}
			assert.NoError(t, oscar.Unmarshal(&flap, bufOut))

			// make sure the sequence number was incremented
			assert.Equal(t, uint32(2), seq)

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
