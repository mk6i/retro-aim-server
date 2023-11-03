package server

import (
	"bytes"
	"context"
	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAlertRouter_RouteAlert(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// input is the request payload
		input XMessage
		// output is the response payload
		output XMessage
		// handlerErr is the mocked handler error response
		handlerErr error
		// expectErr is the expected error returned by the router
		expectErr error
	}{
		{
			name: "receive AlertNotifyCapabilities, return no response",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ALERT,
					SubGroup:  oscar.AlertNotifyCapabilities,
				},
				snacOut: oscar.SnacFrame{},
			},
			output: XMessage{},
		},
		{
			name: "receive AlertNotifyDisplayCapabilities, return no response",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ALERT,
					SubGroup:  oscar.AlertNotifyDisplayCapabilities,
				},
				snacOut: oscar.SnacFrame{},
			},
			output: XMessage{},
		},
		{
			name: "receive AlertGetSubsRequest, expect ErrUnsupportedSubGroup",
			input: XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ALERT,
					SubGroup:  oscar.AlertGetSubsRequest,
				},
				snacOut: struct{}{}, // empty SNAC
			},
			output:    XMessage{}, // empty SNAC
			expectErr: ErrUnsupportedSubGroup,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router := AlertRouter{
				RouteLogger: RouteLogger{
					Logger: NewLogger(Config{}),
				},
			}

			bufIn := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.input.snacOut, bufIn))

			bufOut := &bytes.Buffer{}
			seq := uint32(1)

			err := router.RouteAlert(context.Background(), tc.input.snacFrame)
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}

			if tc.output == (XMessage{}) {
				// make sure no response was sent
				assert.Empty(t, bufOut.Bytes())
				return
			}

			// verify the FLAP frame
			flap := oscar.FlapFrame{}
			assert.NoError(t, oscar.Unmarshal(&flap, bufOut))

			// make sure the sequence number was incremented
			assert.Equal(t, uint32(2), seq)

			flapBuf, err := flap.SNACBuffer(bufOut)
			assert.NoError(t, err)

			// verify the SNAC frame
			snacFrame := oscar.SnacFrame{}
			assert.NoError(t, oscar.Unmarshal(&snacFrame, flapBuf))
			assert.Equal(t, tc.output.snacFrame, snacFrame)

			// verify the SNAC message
			snacBuf := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.output.snacOut, snacBuf))
			assert.Equal(t, snacBuf.Bytes(), flapBuf.Bytes())
		})
	}
}
