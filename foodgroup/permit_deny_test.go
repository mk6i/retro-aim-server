package foodgroup

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mk6i/retro-aim-server/wire"
)

func TestPermitDenyService_RightsQuery(t *testing.T) {
	svc := NewPermitDenyService()

	have := svc.RightsQuery(nil, wire.SNACFrame{RequestID: 1234})
	want := wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.PermitDeny,
			SubGroup:  wire.PermitDenyRightsReply,
			RequestID: 1234,
		},
		Body: wire.SNAC_0x09_0x03_PermitDenyRightsReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLVBE(wire.PermitDenyTLVMaxDenies, uint16(100)),
					wire.NewTLVBE(wire.PermitDenyTLVMaxPermits, uint16(100)),
					wire.NewTLVBE(wire.PermitDenyTLVMaxTempPermits, uint16(100)),
				},
			},
		},
	}

	assert.Equal(t, want, have)
}
