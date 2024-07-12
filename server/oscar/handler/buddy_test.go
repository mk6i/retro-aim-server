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

func TestBuddyHandler_AddBuddies(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Buddy,
			SubGroup:  wire.BuddyAddBuddies,
		},
		Body: wire.SNAC_0x03_0x04_BuddyAddBuddies{
			Buddies: []struct {
				ScreenName string `oscar:"len_prefix=uint8"`
			}{
				{
					ScreenName: "user1",
				},
			},
		},
	}

	svc := newMockBuddyService(t)
	svc.EXPECT().
		AddBuddies(mock.Anything, mock.Anything, input.Body).
		Return(nil)

	h := NewBuddyHandler(slog.Default(), svc)

	responseWriter := newMockResponseWriter(t)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.AddBuddies(nil, nil, input.Frame, buf, responseWriter))
}

func TestBuddyHandler_DelBuddies(t *testing.T) {
	input := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Buddy,
			SubGroup:  wire.BuddyDelBuddies,
		},
		Body: wire.SNAC_0x03_0x05_BuddyDelBuddies{
			Buddies: []struct {
				ScreenName string `oscar:"len_prefix=uint8"`
			}{
				{
					ScreenName: "user1",
				},
			},
		},
	}

	svc := newMockBuddyService(t)
	svc.EXPECT().
		DelBuddies(mock.Anything, mock.Anything, input.Body)

	h := NewBuddyHandler(slog.Default(), svc)

	responseWriter := newMockResponseWriter(t)

	buf := &bytes.Buffer{}
	assert.NoError(t, wire.Marshal(input.Body, buf))

	assert.NoError(t, h.DelBuddies(nil, nil, input.Frame, buf, responseWriter))
}
