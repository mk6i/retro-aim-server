package handler

import (
	"testing"

	"github.com/mk6i/retro-aim-server/oscar"
	"github.com/stretchr/testify/assert"
)

func TestBuddyService_RightsQueryHandler(t *testing.T) {
	svc := NewBuddyService()

	want := oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Buddy,
			SubGroup:  oscar.BuddyRightsReply,
			RequestID: 1234,
		},
		Body: oscar.SNAC_0x03_0x03_BuddyRightsReply{
			TLVRestBlock: oscar.TLVRestBlock{
				TLVList: oscar.TLVList{
					oscar.NewTLV(oscar.BuddyTLVTagsParmMaxBuddies, uint16(100)),
					oscar.NewTLV(oscar.BuddyTLVTagsParmMaxWatchers, uint16(100)),
					oscar.NewTLV(oscar.BuddyTLVTagsParmMaxIcqBroad, uint16(100)),
					oscar.NewTLV(oscar.BuddyTLVTagsParmMaxTempBuddies, uint16(100)),
				},
			},
		},
	}
	have := svc.RightsQueryHandler(nil, oscar.SNACFrame{RequestID: 1234})

	assert.Equal(t, want, have)
}
