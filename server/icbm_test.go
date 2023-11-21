package server

import (
	"bytes"
	"testing"

	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestICBMRouter_RouteICBM(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// input is the request payload
		input oscar.XMessage
		// output is the response payload
		output *oscar.XMessage
		// handlerErr is the mocked handler error response
		handlerErr error
		// expectErr is the expected error returned by the router
		expectErr error
	}{
		{
			name: "receive ICBMAddParameters SNAC, return no response",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMAddParameters,
				},
				SnacOut: oscar.SNAC_0x04_0x02_ICBMAddParameters{
					Channel: 1,
				},
			},
			output: nil,
		},
		{
			name: "receive ICBMParameterQuery, return ICBMParameterReply",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMParameterQuery,
				},
				SnacOut: struct{}{}, // empty SNAC
			},
			output: &oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMParameterReply,
				},
				SnacOut: oscar.SNAC_0x04_0x05_ICBMParameterReply{
					MaxSlots: 100,
				},
			},
		},
		{
			name: "receive ICBMChannelMsgToHost, return ICBMHostAck",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMChannelMsgToHost,
				},
				SnacOut: oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ScreenName: "recipient-screen-name",
				},
			},
			output: &oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMHostAck,
				},
				SnacOut: oscar.SNAC_0x04_0x0C_ICBMHostAck{
					ChannelID: 4,
				},
			},
		},
		{
			name: "receive ICBMChannelMsgToHost, return no reply",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMChannelMsgToHost,
				},
				SnacOut: oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ScreenName: "recipient-screen-name",
				},
			},
			output: nil,
		},
		{
			name: "receive ICBMEvilRequest, return ICBMEvilReply",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMEvilRequest,
				},
				SnacOut: oscar.SNAC_0x04_0x08_ICBMEvilRequest{
					ScreenName: "recipient-screen-name",
				},
			},
			output: &oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMEvilReply,
				},
				SnacOut: oscar.SNAC_0x04_0x09_ICBMEvilReply{
					EvilDeltaApplied: 100,
				},
			},
		},
		{
			name: "receive ICBMClientErr, return no response",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMClientErr,
				},
				SnacOut: oscar.SNAC_0x04_0x0B_ICBMClientErr{
					Code: 4,
				},
			},
			output: nil,
		},
		{
			name: "receive ICBMClientEvent, return no response",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMClientEvent,
				},
				SnacOut: oscar.SNAC_0x04_0x14_ICBMClientEvent{
					ScreenName: "recipient-screen-name",
				},
			},
			output: nil,
		},
		{
			name: "receive ICBMMissedCalls, expect ErrUnsupportedSubGroup",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMMissedCalls,
				},
				SnacOut: struct{}{}, // empty SNAC
			},
			output:    nil,
			expectErr: ErrUnsupportedSubGroup,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newMockICBMHandler(t)
			svc.EXPECT().
				ChannelMsgToHostHandler(mock.Anything, mock.Anything, tc.input.SnacOut).
				Return(tc.output, tc.handlerErr).
				Maybe()
			svc.EXPECT().
				ClientEventHandler(mock.Anything, mock.Anything, tc.input.SnacOut).
				Return(tc.handlerErr).
				Maybe()
			if tc.output != nil {
				svc.EXPECT().
					EvilRequestHandler(mock.Anything, mock.Anything, tc.input.SnacOut).
					Return(*tc.output, tc.handlerErr).
					Maybe()
				svc.EXPECT().
					ParameterQueryHandler(mock.Anything).
					Return(*tc.output).
					Maybe()
			}

			router := ICBMRouter{
				ICBMHandler: svc,
				RouteLogger: RouteLogger{
					Logger: NewLogger(Config{}),
				},
			}

			bufIn := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.input.SnacOut, bufIn))

			bufOut := &bytes.Buffer{}
			seq := uint32(1)

			err := router.RouteICBM(nil, nil, tc.input.SnacFrame, bufIn, bufOut, &seq)
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}

			if tc.output == nil {
				// make sure no response was sent
				assert.Empty(t, bufOut.Bytes())
				return
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
