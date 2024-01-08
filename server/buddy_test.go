package server

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/oscar"
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
							oscar.NewTLV(0x01, uint16(1000)),
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
							oscar.NewTLV(0x01, uint16(1000)),
						},
					},
				},
			},
		},
		{
			name: "receive ChatRowListInfo, return ErrUnsupportedSubGroup",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Chat,
					SubGroup:  oscar.ChatRowListInfo,
				},
				Body: struct{}{},
			},
			expectErr: ErrUnsupportedSubGroup,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newMockBuddyHandler(t)
			svc.EXPECT().
				RightsQueryHandler(mock.Anything, tc.input.Frame).
				Return(tc.output).
				Maybe()

			router := NewBuddyRouter(NewLogger(config.Config{}), svc)

			bufIn := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.input.Body, bufIn))

			bufOut := &bytes.Buffer{}
			seq := uint32(0)

			err := router.RouteBuddy(nil, tc.input.Frame, bufIn, bufOut, &seq)
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}

			// verify the FLAP frame
			flap := oscar.FLAPFrame{}
			assert.NoError(t, oscar.Unmarshal(&flap, bufOut))

			// make sure the sequence increments
			assert.Equal(t, seq, uint32(1))
			assert.Equal(t, flap.Sequence, uint16(0))

			flapBuf, err := flap.ReadBody(bufOut)
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

func TestNewBuddyRouter(t *testing.T) {
	type args struct {
		logger       *slog.Logger
		buddyHandler BuddyHandler
	}
	tests := []struct {
		name string
		args args
		want BuddyRouter
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewBuddyRouter(tt.args.logger, tt.args.buddyHandler), "NewBuddyRouter(%v, %v)", tt.args.logger, tt.args.buddyHandler)
		})
	}
}
