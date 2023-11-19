package server

import (
	"bytes"
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
)

func TestChatRouter_RouteChat(t *testing.T) {
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
			name: "receive ChatChannelMsgToHost, return ChatChannelMsgToClient",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.CHAT,
					SubGroup:  oscar.ChatChannelMsgToHost,
				},
				SnacOut: oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost{
					Channel: 4,
				},
			},
			output: &oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.CHAT,
					SubGroup:  oscar.ChatChannelMsgToClient,
				},
				SnacOut: oscar.SNAC_0x0E_0x06_ChatChannelMsgToClient{
					Channel: 4,
				},
			},
		},
		{
			name: "receive ChatChannelMsgToHost, return no response",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.CHAT,
					SubGroup:  oscar.ChatChannelMsgToHost,
				},
				SnacOut: oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost{
					Channel: 4,
				},
			},
			output: nil,
		},
		{
			name: "receive ChatRowListInfo, return ErrUnsupportedSubGroup",
			input: oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.CHAT,
					SubGroup:  oscar.ChatRowListInfo,
				},
				SnacOut: struct{}{},
			},
			output:    nil,
			expectErr: ErrUnsupportedSubGroup,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewMockChatHandler(t)
			svc.EXPECT().
				ChannelMsgToHostHandler(mock.Anything, mock.Anything, mock.Anything, tc.input.SnacOut).
				Return(tc.output, tc.handlerErr).
				Maybe()

			router := ChatRouter{
				ChatHandler: svc,
				RouteLogger: RouteLogger{
					Logger: NewLogger(Config{}),
				},
			}

			bufIn := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.input.SnacOut, bufIn))

			bufOut := &bytes.Buffer{}
			seq := uint32(0)

			err := router.RouteChat(nil, nil, nil, tc.input.SnacFrame, bufIn, bufOut, &seq)
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}

			if tc.output == nil {
				// make sure no response was sent
				assert.Empty(t, bufOut.Bytes())
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
