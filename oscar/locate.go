package oscar

import (
	"fmt"
	"io"
	"reflect"
)

const (
	LocateErr                  uint16 = 0x0001
	LocateRightsQuery                 = 0x0002
	LocateRightsReply                 = 0x0003
	LocateSetInfo                     = 0x0004
	LocateUserInfoQuery               = 0x0005
	LocateUserInfoReply               = 0x0006
	LocateWatcherSubRequest           = 0x0007
	LocateWatcherNotification         = 0x0008
	LocateSetDirInfo                  = 0x0009
	LocateSetDirReply                 = 0x000A
	LocateGetDirInfo                  = 0x000B
	LocateGetDirReply                 = 0x000C
	LocateGroupCapabilityQuery        = 0x000D
	LocateGroupCapabilityReply        = 0x000E
	LocateSetKeywordInfo              = 0x000F
	LocateSetKeywordReply             = 0x0010
	LocateGetKeywordInfo              = 0x0011
	LocateGetKeywordReply             = 0x0012
	LocateFindListByEmail             = 0x0013
	LocateFindListReply               = 0x0014
	LocateUserInfoQuery2              = 0x0015
)

func routeLocate(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence uint16) error {
	switch snac.subGroup {
	case LocateErr:
		panic("not implemented")
	case LocateRightsQuery:
		return SendAndReceiveLocateRights(flap, snac, r, w, sequence)
	case LocateSetInfo:
		return ReceiveSetInfo(flap, snac, r, w, sequence)
	case LocateUserInfoQuery:
		panic("not implemented")
	case LocateUserInfoReply:
		panic("not implemented")
	case LocateWatcherSubRequest:
		panic("not implemented")
	case LocateWatcherNotification:
		panic("not implemented")
	case LocateSetDirInfo:
		panic("not implemented")
	case LocateSetDirReply:
		panic("not implemented")
	case LocateGetDirInfo:
		panic("not implemented")
	case LocateGetDirReply:
		panic("not implemented")
	case LocateGroupCapabilityQuery:
		panic("not implemented")
	case LocateGroupCapabilityReply:
		panic("not implemented")
	case LocateSetKeywordInfo:
		panic("not implemented")
	case LocateSetKeywordReply:
		panic("not implemented")
	case LocateGetKeywordInfo:
		panic("not implemented")
	case LocateGetKeywordReply:
		panic("not implemented")
	case LocateFindListByEmail:
		panic("not implemented")
	case LocateFindListReply:
		panic("not implemented")
	case LocateUserInfoQuery2:
		panic("not implemented")
	}

	return nil
}

type snacLocateRightsReply struct {
	TLVPayload
}

func (s *snacLocateRightsReply) write(w io.Writer) error {
	return s.TLVPayload.write(w)
}

func SendAndReceiveLocateRights(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence uint16) error {
	fmt.Printf("sendAndReceiveLocateRights read SNAC frame: %+v\n", snac)

	snacFrameOut := snacFrame{
		foodGroup: LOCATE,
		subGroup:  LocateRightsReply,
	}
	snacPayloadOut := &snacLocateRightsReply{
		TLVPayload: TLVPayload{
			TLVs: []*TLV{
				{
					tType: 0x01,
					val:   uint16(1000),
				},
				{
					tType: 0x02,
					val:   uint16(1000),
				},
				{
					tType: 0x03,
					val:   uint16(1000),
				},
				{
					tType: 0x04,
					val:   uint16(1000),
				},
				{
					tType: 0x05,
					val:   uint16(1000),
				},
			},
		},
	}

	return writeOutSNAC(flap, snacFrameOut, snacPayloadOut, sequence, w)
}

func ReceiveSetInfo(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence uint16) error {
	fmt.Printf("ReceiveSetInfo read SNAC frame: %+v\n", snac)

	snacPayload := &TLVPayload{}
	lookup := map[uint16]reflect.Kind{
		0x05: reflect.Slice,
	}
	if err := snacPayload.read(r, lookup); err != nil {
		return err
	}

	fmt.Printf("ReceiveSetInfo read SNAC: %+v\n", snacPayload)

	return nil
}
