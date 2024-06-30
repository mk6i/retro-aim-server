package foodgroup

import (
	"context"

	"github.com/mk6i/retro-aim-server/wire"
)

// NewICQService creates an instance of ICQService.
func NewICQService() ICQService {
	return ICQService{}
}

// ICQService provides functionality for the ICQ (PD) food group.
// The PD food group manages settings for permit/deny (allow/block) for
// pre-feedbag (sever-side buddy list) AIM clients. Right now it's stubbed out
// to support pidgin. Eventually this food group will be fully implemented in
// order to support client blocking in AIM <= 3.0.
type ICQService struct {
}

func (s ICQService) DBQuery(_ context.Context, frame wire.SNACFrame, body wire.SNAC_0x0F_0x02_ICQDBQuery) wire.SNACMessage {
	return wire.SNACMessage{}
	//return wire.SNACMessage{
	//	Frame: wire.SNACFrame{
	//		FoodGroup: wire.ICQ,
	//		SubGroup:  wire.ICQRightsReply,
	//		RequestID: frame.RequestID,
	//	},
	//	Body: wire.SNAC_0x09_0x03_ICQRightsReply{
	//		TLVRestBlock: wire.TLVRestBlock{
	//			TLVList: wire.TLVList{
	//				wire.NewTLV(wire.ICQTLVMaxDenies, uint16(100)),
	//				wire.NewTLV(wire.ICQTLVMaxPermits, uint16(100)),
	//				wire.NewTLV(wire.ICQTLVMaxTempPermits, uint16(100)),
	//			},
	//		},
	//	},
	//}
}
