package oscar

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
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

func routeLocate(sess *Session, sm *SessionManager, fm *FeedbagStore, snac snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch snac.subGroup {
	case LocateErr:
		panic("not implemented")
	case LocateRightsQuery:
		return SendAndReceiveLocateRights(snac, w, sequence)
	case LocateSetInfo:
		return ReceiveSetInfo(sess, sm, fm, snac, r)
	case LocateUserInfoQuery:
		panic("not implemented")
	case LocateUserInfoReply:
		panic("not implemented")
	case LocateWatcherSubRequest:
		panic("not implemented")
	case LocateWatcherNotification:
		panic("not implemented")
	case LocateSetDirInfo:
		return SendAndReceiveSetDirInfo(snac, r, w, sequence)
	case LocateGetDirInfo:
		return ReceiveLocateGetDirInfo(snac, r)
	case LocateGetDirReply:
		panic("not implemented")
	case LocateGroupCapabilityQuery:
		panic("not implemented")
	case LocateGroupCapabilityReply:
		panic("not implemented")
	case LocateSetKeywordInfo:
		return SendAndReceiveSetKeywordInfo(snac, r, w, sequence)
	case LocateGetKeywordInfo:
		panic("not implemented")
	case LocateGetKeywordReply:
		panic("not implemented")
	case LocateFindListByEmail:
		panic("not implemented")
	case LocateFindListReply:
		panic("not implemented")
	case LocateUserInfoQuery2:
		return SendAndReceiveUserInfoQuery2(sess, sm, fm, snac, r, w, sequence)
	}

	return nil
}

type snacLocateRightsReply struct {
	TLVPayload
}

func (s snacLocateRightsReply) write(w io.Writer) error {
	return s.TLVPayload.write(w)
}

func SendAndReceiveLocateRights(snac snacFrame, w io.Writer, sequence *uint32) error {
	fmt.Printf("sendAndReceiveLocateRights read SNAC frame: %+v\n", snac)

	snacFrameOut := snacFrame{
		foodGroup: LOCATE,
		subGroup:  LocateRightsReply,
	}
	snacPayloadOut := snacLocateRightsReply{
		TLVPayload: TLVPayload{
			TLVs: []TLV{
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

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
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

func ReceiveSetInfo(sess *Session, sm *SessionManager, fm *FeedbagStore, snac snacFrame, r io.Reader) error {
	fmt.Printf("ReceiveSetInfo read SNAC frame: %+v\n", snac)

	snacPayload := TLVPayload{}
	if err := snacPayload.read(r); err != nil {
		return err
	}

	// update profile
	if profile, hasProfile := snacPayload.getString(LocateTlvTagsInfoSigData); hasProfile {
		if err := fm.UpsertProfile(sess.ScreenName, profile); err != nil {
			return err
		}
	}

	// broadcast away message change to buddies
	if awayMsg, hasAwayMsg := snacPayload.getString(LocateTlvTagsInfoUnavailableData); hasAwayMsg {
		sess.SetAwayMessage(awayMsg)
		if err := NotifyArrival(sess, sm, fm); err != nil {
			return err
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

func ReceiveLocateGetDirInfo(snac snacFrame, r io.Reader) error {
	fmt.Printf("ReceiveLocateGetDirInfo read SNAC frame: %+v\n", snac)

	snacPayload := snacDirInfo{}
	if err := snacPayload.read(r); err != nil {
		return err
	}

	fmt.Printf("ReceiveLocateGetDirInfo read SNAC: %+v\n", snacPayload)

	return nil
}

type snacUserInfoQuery2 struct {
	type2      uint32
	screenName string
}

func (s *snacUserInfoQuery2) read(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &s.type2); err != nil {
		return err
	}
	var l uint8
	if err := binary.Read(r, binary.BigEndian, &l); err != nil {
		return err
	}
	buf := make([]byte, l)
	if _, err := r.Read(buf); err != nil {
		return err
	}
	s.screenName = string(buf)
	return nil
}

type snacUserInfoReply struct {
	screenName    string
	warningLevel  uint16
	userInfo      TLVPayload
	clientProfile TLVPayload
	awayMessage   TLVPayload
}

func (f snacUserInfoReply) write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, uint8(len(f.screenName))); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, []byte(f.screenName)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, f.warningLevel); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint16(len(f.userInfo.TLVs))); err != nil {
		return err
	}
	if err := f.userInfo.write(w); err != nil {
		return err
	}
	if err := f.clientProfile.write(w); err != nil {
		return err
	}
	return f.awayMessage.write(w)
}

func SendAndReceiveUserInfoQuery2(sess *Session, sm *SessionManager, fm *FeedbagStore, snac snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("SendAndReceiveUserInfoQuery2 read SNAC frame: %+v\n", snac)

	snacPayloadIn := snacUserInfoQuery2{}
	if err := snacPayloadIn.read(r); err != nil {
		return err
	}

	blocked, err := fm.Blocked(sess.ScreenName, snacPayloadIn.screenName)
	if err != nil {
		return err
	}
	if blocked != BlockedNo {
		snacFrameOut := snacFrame{
			foodGroup: LOCATE,
			subGroup:  LocateErr,
		}
		snacPayloadOut := snacError{
			code: ErrorCodeNotLoggedOn,
		}
		return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
	}

	buddySess, err := sm.RetrieveByScreenName(snacPayloadIn.screenName)
	if err != nil {
		if errors.Is(err, errSessNotFound) {
			snacFrameOut := snacFrame{
				foodGroup: LOCATE,
				subGroup:  LocateErr,
			}
			snacPayloadOut := snacError{
				code: ErrorCodeNotLoggedOn,
			}
			return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
		}
	}

	snacFrameOut := snacFrame{
		foodGroup: LOCATE,
		subGroup:  LocateUserInfoReply,
	}
	snacPayloadOut := snacUserInfoReply{
		screenName:   snacPayloadIn.screenName,
		warningLevel: buddySess.GetWarning(),
		userInfo: TLVPayload{
			TLVs: buddySess.GetUserInfo(),
		},
		clientProfile: TLVPayload{},
		awayMessage:   TLVPayload{},
	}

	// profile
	if snacPayloadIn.type2&1 == 1 {
		profile, err := fm.RetrieveProfile(snacPayloadIn.screenName)
		if err != nil {
			if err == errUserNotExist {
				snacFrameOut := snacFrame{
					foodGroup: LOCATE,
					subGroup:  LocateErr,
				}
				snacPayloadOut := snacError{
					code: ErrorCodeNotLoggedOn,
				}
				return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
			}
			return err
		}
		snacPayloadOut.clientProfile.TLVs = append(snacPayloadOut.clientProfile.TLVs, []TLV{
			{
				tType: 0x01,
				val:   `text/aolrtf; charset="us-ascii"`,
			},
			{
				tType: 0x02,
				val:   profile,
			},
		}...)
	}

	// away message
	if snacPayloadIn.type2&2 == 2 {
		snacPayloadOut.clientProfile.TLVs = append(snacPayloadOut.clientProfile.TLVs, []TLV{
			{
				tType: 0x03,
				val:   `text/aolrtf; charset="us-ascii"`,
			},
			{
				tType: 0x04,
				val:   buddySess.GetAwayMessage(),
			},
		}...)
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

type snacSetDirInfo struct {
	TLVPayload
}

func (s *snacSetDirInfo) read(r io.Reader) error {
	return s.TLVPayload.read(r)
}

type snacSetDirInfoReply struct {
	result uint16
}

func (s snacSetDirInfoReply) write(w io.Writer) error {
	return binary.Write(w, binary.BigEndian, s.result)
}

func SendAndReceiveSetDirInfo(snac snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("SendAndReceiveSetDirInfo read SNAC frame: %+v\n", snac)

	snacPayloadIn := snacSetDirInfo{}
	if err := snacPayloadIn.read(r); err != nil {
		return err
	}

	snacFrameOut := snacFrame{
		foodGroup: LOCATE,
		subGroup:  LocateSetDirReply,
	}
	snacPayloadOut := snacSetDirInfoReply{
		result: 1,
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

type snacSetKeywordInfo struct {
	payload []byte
}

func (s *snacSetKeywordInfo) read(r io.Reader) error {
	_, err := r.Read(s.payload)
	return err
}

type snacSetKeywordInfoReply struct {
	unknown uint16
}

func (s snacSetKeywordInfoReply) write(w io.Writer) error {
	return binary.Write(w, binary.BigEndian, s.unknown)
}

func SendAndReceiveSetKeywordInfo(snac snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("SendAndReceiveSetKeywordInfo read SNAC frame: %+v\n", snac)

	snacPayloadIn := snacSetKeywordInfo{}
	if err := snacPayloadIn.read(r); err != nil {
		return err
	}

	snacFrameOut := snacFrame{
		foodGroup: LOCATE,
		subGroup:  LocateSetKeywordReply,
	}
	snacPayloadOut := snacSetKeywordInfoReply{
		unknown: 1,
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}
