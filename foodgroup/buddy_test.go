package foodgroup

import (
	"testing"

	"github.com/mk6i/retro-aim-server/wire"

	"github.com/stretchr/testify/assert"
)

func TestBuddyService_RightsQuery(t *testing.T) {
	svc := NewBuddyService()

	want := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.Buddy,
			SubGroup:  wire.BuddyRightsReply,
			RequestID: 1234,
		},
		Body: wire.SNAC_0x03_0x03_BuddyRightsReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.BuddyTLVTagsParmMaxBuddies, uint16(100)),
					wire.NewTLV(wire.BuddyTLVTagsParmMaxWatchers, uint16(100)),
					wire.NewTLV(wire.BuddyTLVTagsParmMaxIcqBroad, uint16(100)),
					wire.NewTLV(wire.BuddyTLVTagsParmMaxTempBuddies, uint16(100)),
				},
			},
		},
	}
	have := svc.RightsQuery(nil, wire.SNACFrame{RequestID: 1234})

	assert.Equal(t, want, have)
}
