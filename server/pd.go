package server

import (
	"fmt"
	"github.com/mkaminski/goaim/oscar"
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

func routePD(snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch snac.SubGroup {
	case PDRightsQuery:
		return SendAndReceivePDRightsQuery(snac, r, w, sequence)
	default:
		return handleUnimplementedSNAC(snac, w, sequence)
	}
}

func SendAndReceivePDRightsQuery(snac oscar.SnacFrame, _ io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("sendAndReceivePDRightsQuery read SNAC frame: %+v\n", snac)

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: PD,
		SubGroup:  PDRightsReply,
	}
	snacPayloadOut := oscar.SNAC_0x09_0x03_PDRightsReply{
		TLVRestBlock: oscar.TLVRestBlock{
			TLVList: oscar.TLVList{
				{
					TType: 0x01,
					Val:   uint16(100),
				},
				{
					TType: 0x02,
					Val:   uint16(100),
				},
				{
					TType: 0x03,
					Val:   uint16(100),
				},
			},
		},
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}
