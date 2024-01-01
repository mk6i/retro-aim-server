package server

import (
	"bytes"
	"testing"

	"github.com/mk6i/retro-aim-server/oscar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestChatNavRouter_RouteChatNavRouter(t *testing.T) {
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
			name: "receive ChatNavRequestChatRights, return ChatNavNavInfo",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ChatNav,
					SubGroup:  oscar.ChatNavRequestChatRights,
				},
				Body: struct{}{},
			},
			output: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ChatNav,
					SubGroup:  oscar.ChatNavNavInfo,
				},
				Body: oscar.SNAC_0x0D_0x09_ChatNavNavInfo{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(0x02, uint8(10)),
						},
					},
				},
			},
		},
		{
			name: "receive ChatNavRequestRoomInfo, return ChatNavNavInfo",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ChatNav,
					SubGroup:  oscar.ChatNavRequestRoomInfo,
				},
				Body: oscar.SNAC_0x0D_0x04_ChatNavRequestRoomInfo{
					Exchange: 1,
				},
			},
			output: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ChatNav,
					SubGroup:  oscar.ChatNavNavInfo,
				},
				Body: oscar.SNAC_0x0D_0x09_ChatNavNavInfo{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(0x02, uint8(10)),
						},
					},
				},
			},
		},
		{
			name: "receive ChatNavCreateRoom, return ChatNavNavInfo",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ChatNav,
					SubGroup:  oscar.ChatNavCreateRoom,
				},
				Body: oscar.SNAC_0x0E_0x02_ChatRoomInfoUpdate{
					Exchange: 1,
				},
			},
			output: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ChatNav,
					SubGroup:  oscar.ChatNavNavInfo,
				},
				Body: oscar.SNAC_0x0D_0x09_ChatNavNavInfo{
					TLVRestBlock: oscar.TLVRestBlock{
						TLVList: oscar.TLVList{
							oscar.NewTLV(0x02, uint8(10)),
						},
					},
				},
			},
		},
		{
			name: "receive ChatNavRequestOccupantList, return ErrUnsupportedSubGroup",
			input: oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.ChatNav,
					SubGroup:  oscar.ChatNavRequestOccupantList,
				},
				Body: struct{}{},
			},
			output:    oscar.SNACMessage{},
			expectErr: ErrUnsupportedSubGroup,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newMockChatNavHandler(t)
			svc.EXPECT().
				RequestChatRightsHandler(mock.Anything, tc.input.Frame).
				Return(tc.output).
				Maybe()
			svc.EXPECT().
				RequestRoomInfoHandler(mock.Anything, tc.input.Frame, tc.input.Body).
				Return(tc.output, tc.handlerErr).
				Maybe()
			svc.EXPECT().
				CreateRoomHandler(mock.Anything, mock.Anything, tc.input.Frame, tc.input.Body).
				Return(tc.output, tc.handlerErr).
				Maybe()

			router := NewChatNavRouter(svc, NewLogger(Config{}))

			bufIn := &bytes.Buffer{}
			assert.NoError(t, oscar.Marshal(tc.input.Body, bufIn))

			bufOut := &bytes.Buffer{}
			seq := uint32(0)

			err := router.RouteChatNav(nil, nil, tc.input.Frame, bufIn, bufOut, &seq)
			assert.ErrorIs(t, err, tc.expectErr)
			if tc.expectErr != nil {
				return
			}

			if tc.output.Frame == (oscar.SNACFrame{}) {
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
