package server

import (
	"bytes"
	"context"
	"testing"

	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
)

func TestAlertRouter_RouteAlert(t *testing.T) {
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
			name: "receive AlertNotifyCapabilities, return no response",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Alert,
					SubGroup:  oscar.AlertNotifyCapabilities,
				},
				Body: oscar.SNACFrame{},
			},
			output: oscar.SNACMessage{},
		},
		{
			name: "receive AlertNotifyDisplayCapabilities, return no response",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Alert,
					SubGroup:  oscar.AlertNotifyDisplayCapabilities,
				},
				Body: oscar.SNACFrame{},
			},
			output: oscar.SNACMessage{},
		},
		{
			name: "receive AlertGetSubsRequest, expect ErrUnsupportedSubGroup",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Alert,
					SubGroup:  oscar.AlertGetSubsRequest,
				},
				Body: struct{}{}, // empty SNAC
			},
			output:    oscar.SNACMessage{}, // empty SNAC
			expectErr: ErrUnsupportedSubGroup,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router := NewAlertRouter(NewLogger(Config{}))

			bufIn := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.input.Body, bufIn))

			bufOut := &bytes.Buffer{}
			seq := uint32(1)

			err := router.RouteAlert(context.Background(), tc.input.Frame)
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
