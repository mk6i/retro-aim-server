package oscar

import (
	"fmt"
	"io"
)

const (
	PDErr                      uint16 = 0x0001
	PDRightsQuery                     = 0x0002
	PDRightsReply                     = 0x0003
	PDSetGroupPermitMask              = 0x0004
	PDAddPermListEntries              = 0x0005
	PDDelPermListEntries              = 0x0006
	PDAddDenyListEntries              = 0x0007
	PDDelDenyListEntries              = 0x0008
	PDBosErr                          = 0x0009
	PDAddTempPermitListEntries        = 0x000A
	PDDelTempPermitListEntries        = 0x000B
)

func routePD(snac snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch snac.subGroup {
	case PDErr:
		panic("not implemented")
	case PDRightsQuery:
		return SendAndReceivePDRightsQuery(snac, r, w, sequence)
	case PDSetGroupPermitMask:
		panic("not implemented")
	case PDAddPermListEntries:
		panic("not implemented")
	case PDDelPermListEntries:
		panic("not implemented")
	case PDAddDenyListEntries:
		panic("not implemented")
	case PDDelDenyListEntries:
		panic("not implemented")
	case PDBosErr:
		panic("not implemented")
	case PDAddTempPermitListEntries:
		panic("not implemented")
	case PDDelTempPermitListEntries:
		panic("not implemented")
	}

	return nil
}

func SendAndReceivePDRightsQuery(snac snacFrame, _ io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("sendAndReceivePDRightsQuery read SNAC frame: %+v\n", snac)

	snacFrameOut := snacFrame{
		foodGroup: PD,
		subGroup:  PDRightsReply,
	}
	snacPayloadOut := TLVRestBlock{
		TLVList: TLVList{
			{
				tType: 0x01,
				val:   uint16(100),
			},
			{
				tType: 0x02,
				val:   uint16(100),
			},
			{
				tType: 0x03,
				val:   uint16(100),
			},
		},
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}
