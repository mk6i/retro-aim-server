package server

import (
	"bytes"
	"testing"

	"github.com/mk6i/retro-aim-server/oscar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestICBMRouter_RouteICBM(t *testing.T) {
	cases := []struct {
		// name is the unit test name
		name string
		// input is the request payload
		input oscar.SNACMessage
		// output is the response payload
		output *oscar.SNACMessage
		// handlerErr is the mocked handler error response
		handlerErr error
		// expectErr is the expected error returned by the router
		expectErr error
	}{
		{
			name: "receive ICBMAddParameters SNAC, return no response",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMAddParameters,
				},
				Body: oscar.SNAC_0x04_0x02_ICBMAddParameters{
					Channel: 1,
				},
			},
			output: nil,
		},
		{
			name: "receive ICBMParameterQuery, return ICBMParameterReply",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMParameterQuery,
				},
				Body: struct{}{}, // empty SNAC
			},
			output: &oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMParameterReply,
				},
				Body: oscar.SNAC_0x04_0x05_ICBMParameterReply{
					MaxSlots: 100,
				},
			},
		},
		{
			name: "receive ICBMChannelMsgToHost, return ICBMHostAck",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMChannelMsgToHost,
				},
				Body: oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ScreenName: "recipient-screen-name",
				},
			},
			output: &oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMHostAck,
				},
				Body: oscar.SNAC_0x04_0x0C_ICBMHostAck{
					ChannelID: 4,
				},
			},
		},
		{
			name: "receive ICBMChannelMsgToHost, return no reply",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMChannelMsgToHost,
				},
				Body: oscar.SNAC_0x04_0x06_ICBMChannelMsgToHost{
					ScreenName: "recipient-screen-name",
				},
			},
			output: nil,
		},
		{
			name: "receive ICBMEvilRequest, return ICBMEvilReply",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMEvilRequest,
				},
				Body: oscar.SNAC_0x04_0x08_ICBMEvilRequest{
					ScreenName: "recipient-screen-name",
				},
			},
			output: &oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMEvilReply,
				},
				Body: oscar.SNAC_0x04_0x09_ICBMEvilReply{
					EvilDeltaApplied: 100,
				},
			},
		},
		{
			name: "receive ICBMClientErr, return no response",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMClientErr,
				},
				Body: oscar.SNAC_0x04_0x0B_ICBMClientErr{
					Code: 4,
				},
			},
			output: nil,
		},
		{
			name: "receive ICBMClientEvent, return no response",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMClientEvent,
				},
				Body: oscar.SNAC_0x04_0x14_ICBMClientEvent{
					ScreenName: "recipient-screen-name",
				},
			},
			output: nil,
		},
		{
			name: "receive ICBMMissedCalls, expect ErrUnsupportedSubGroup",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ICBM,
					SubGroup:  oscar.ICBMMissedCalls,
				},
				Body: struct{}{}, // empty SNAC
			},
			output:    nil,
			expectErr: ErrUnsupportedSubGroup,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newMockICBMHandler(t)
			svc.EXPECT().
				ChannelMsgToHostHandler(mock.Anything, mock.Anything, tc.input.Frame, tc.input.Body).
				Return(tc.output, tc.handlerErr).
				Maybe()
			svc.EXPECT().
				ClientEventHandler(mock.Anything, mock.Anything, tc.input.Frame, tc.input.Body).
				Return(tc.handlerErr).
				Maybe()
			if tc.output != nil {
				svc.EXPECT().
					EvilRequestHandler(mock.Anything, mock.Anything, tc.input.Frame, tc.input.Body).
					Return(*tc.output, tc.handlerErr).
					Maybe()
				svc.EXPECT().
					ParameterQueryHandler(mock.Anything, tc.input.Frame).
					Return(*tc.output).
					Maybe()
			}

			router := NewICBMRouter(NewLogger(Config{}), svc)

			bufIn := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.input.Body, bufIn))

			bufOut := &bytes.Buffer{}
			seq := uint32(1)

			err := router.RouteICBM(nil, nil, tc.input.Frame, bufIn, bufOut, &seq)
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
