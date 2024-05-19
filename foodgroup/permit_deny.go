package foodgroup

import (
	"context"

	"github.com/mk6i/retro-aim-server/wire"
)

// NewPermitDenyService creates an instance of PermitDenyService.
func NewPermitDenyService() PermitDenyService {
	return PermitDenyService{}
}

// PermitDenyService provides functionality for the PermitDeny (PD) food group.
// The PD food group manages settings for permit/deny (allow/block) for
// pre-feedbag (sever-side buddy list) AIM clients. Right now it's stubbed out
// to support pidgin. Eventually this food group will be fully implemented in
// order to support client blocking in AIM <= 3.0.
type PermitDenyService struct {
}

// RightsQuery returns settings for the PermitDeny food group. It returns SNAC
// wire.PermitDenyRightsReply. The values in the return SNAC were arbitrarily
// chosen.
func (s PermitDenyService) RightsQuery(_ context.Context, frame wire.SNACFrame) wire.SNACMessage {
	return wire.SNACMessage{
		Frame: wire.SNACFrame{
			FoodGroup: wire.PermitDeny,
			SubGroup:  wire.PermitDenyRightsReply,
			RequestID: frame.RequestID,
		},
		Body: wire.SNAC_0x09_0x03_PermitDenyRightsReply{
			TLVRestBlock: wire.TLVRestBlock{
				TLVList: wire.TLVList{
					wire.NewTLV(wire.PermitDenyTLVMaxDenies, uint16(100)),
					wire.NewTLV(wire.PermitDenyTLVMaxPermits, uint16(100)),
					wire.NewTLV(wire.PermitDenyTLVMaxTempPermits, uint16(100)),
				},
			},
		},
	}
}
