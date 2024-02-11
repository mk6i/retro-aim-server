package handler

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBuddyHandler_RightsQuery(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Buddy,
			SubGroup:  wire.BuddyRightsQuery,
		},
		Body: wire.SNAC_0x03_0x02_BuddyRightsQuery{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(0x01, uint16(1000)),
				},
			},
		},
	}
	output := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Buddy,
			SubGroup:  wire.BuddyRightsReply,
		},
		Body: wire.SNAC_0x03_0x03_BuddyRightsReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(0x01, uint16(1000)),
				},
			},
		},
	}

	svc := newMockBuddyService(t)
	svc.EXPECT().
		RightsQuery(mock.Anything, input.Frame).
		Return(output)

	h := NewBuddyHandler(slog.Default(), svc)

	responseWriter := newMockResponseWriter(t)
	responseWriter.EXPECT().
		SendSNAC(output.Frame, output.Body).
		Return(nil)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.RightsQuery(nil, nil, input.Frame, buf, responseWriter))
}
