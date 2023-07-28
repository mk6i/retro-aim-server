package oscar

import (
	"encoding/binary"
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

func routeLocate(sess *Session, sm *SessionManager, fm *FeedbagStore, flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch snac.subGroup {
	case LocateErr:
		panic("not implemented")
	case LocateRightsQuery:
		return SendAndReceiveLocateRights(flap, snac, r, w, sequence)
	case LocateSetInfo:
		return ReceiveSetInfo(sess, sm, fm, flap, snac, r, w, sequence)
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
		return ReceiveLocateGetDirInfo(flap, snac, r, w, sequence)
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

func SendAndReceiveLocateRights(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
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

	return writeOutSNAC(snac, flap, snacFrameOut, snacPayloadOut, sequence, w)
}

var (
	LocateTlvTagsInfoSigMime         = uint16(0x01)
	LocateTlvTagsInfoSigData         = uint16(0x02)
	LocateTlvTagsInfoUnavailableMime = uint16(0x03)
	LocateTlvTagsInfoUnavailableData = uint16(0x04)
	LocateTlvTagsInfoCapabilities    = uint16(0x05)
	LocateTlvTagsInfoCerts           = uint16(0x06)
	LocateTlvTagsInfoSigTime         = uint16(0x0A)
	LocateTlvTagsInfoUnavailableTime = uint16(0x0B)
	LocateTlvTagsInfoSupportHostSig  = uint16(0x0C)
	LocateTlvTagsInfoHtmlInfoData    = uint16(0x0E)
	LocateTlvTagsInfoHtmlInfoType    = uint16(0x0D)
)

func ReceiveSetInfo(sess *Session, sm *SessionManager, fm *FeedbagStore, flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("ReceiveSetInfo read SNAC frame: %+v\n", snac)

	snacPayload := &TLVPayload{}
	lookup := map[uint16]reflect.Kind{
		LocateTlvTagsInfoSigMime:         reflect.String,
		LocateTlvTagsInfoSigData:         reflect.String,
		LocateTlvTagsInfoUnavailableMime: reflect.String,
		LocateTlvTagsInfoUnavailableData: reflect.String,
		LocateTlvTagsInfoCapabilities:    reflect.Slice,
		LocateTlvTagsInfoCerts:           reflect.Uint32,
		LocateTlvTagsInfoSigTime:         reflect.Uint32,
		LocateTlvTagsInfoUnavailableTime: reflect.Uint32,
		LocateTlvTagsInfoSupportHostSig:  reflect.Uint8,
		LocateTlvTagsInfoHtmlInfoType:    reflect.String,
		LocateTlvTagsInfoHtmlInfoData:    reflect.String,
	}
	if err := snacPayload.read(r, lookup); err != nil {
		return err
	}

	// broadcast away message change to buddies
	if awayMsg, hasAwayMsg := snacPayload.getString(LocateTlvTagsInfoUnavailableData); hasAwayMsg {
		if awayMsg != "" {
			if err := NotifyAway(sess, sm, fm, awayMsg); err != nil {
				return err
			}
		} else {
			// clear array message
			if err := NotifyArrival(sess, sm, fm); err != nil {
				return err
			}
		}
	}

	fmt.Printf("ReceiveSetInfo read SNAC: %+v\n", snacPayload)

	return nil
}

type snacDirInfo struct {
	watcherScreenNames []string
}

func (s *snacDirInfo) read(r io.Reader) error {
	for {
		var l uint8
		if err := binary.Read(r, binary.BigEndian, &l); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		buf := make([]byte, l)
		if _, err := r.Read(buf); err != nil {
			return err
		}
		s.watcherScreenNames = append(s.watcherScreenNames, string(buf))
	}
	return nil
}

func ReceiveLocateGetDirInfo(flap *flapFrame, snac *snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("ReceiveLocateGetDirInfo read SNAC frame: %+v\n", snac)

	snacPayload := &snacDirInfo{}
	if err := snacPayload.read(r); err != nil {
		return err
	}

	fmt.Printf("ReceiveLocateGetDirInfo read SNAC: %+v\n", snacPayload)

	return nil
}
