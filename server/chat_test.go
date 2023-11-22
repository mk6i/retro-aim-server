package server

import (
	"bytes"
	"testing"

	"github.com/mkaminski/goaim/oscar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestChatRouter_RouteChat(t *testing.T) {
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
			name: "receive ChatChannelMsgToHost, return ChatChannelMsgToClient",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Chat,
					SubGroup:  oscar.ChatChannelMsgToHost,
				},
				Body: oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost{
					Channel: 4,
				},
			},
			output: &oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Chat,
					SubGroup:  oscar.ChatChannelMsgToClient,
				},
				Body: oscar.SNAC_0x0E_0x06_ChatChannelMsgToClient{
					Channel: 4,
				},
			},
		},
		{
			name: "receive ChatChannelMsgToHost, return no response",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Chat,
					SubGroup:  oscar.ChatChannelMsgToHost,
				},
				Body: oscar.SNAC_0x0E_0x05_ChatChannelMsgToHost{
					Channel: 4,
				},
			},
			output: nil,
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
			output:    nil,
			expectErr: ErrUnsupportedSubGroup,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newMockChatHandler(t)
			svc.EXPECT().
				ChannelMsgToHostHandler(mock.Anything, mock.Anything, mock.Anything, tc.input.Body).
				Return(tc.output, tc.handlerErr).
				Maybe()

			router := ChatRouter{
				ChatHandler: svc,
				RouteLogger: RouteLogger{
					Logger: NewLogger(Config{}),
				},
			}

			bufIn := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.input.Body, bufIn))

			bufOut := &bytes.Buffer{}
			seq := uint32(0)

			err := router.RouteChat(nil, nil, "", tc.input.Frame, bufIn, bufOut, &seq)
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
