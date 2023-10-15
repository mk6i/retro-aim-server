package server

import (
	"errors"
	"fmt"
	"github.com/mkaminski/goaim/oscar"
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

const (
	LocateTLVTagsInfoSigMime         uint16 = 0x01
	LocateTLVTagsInfoSigData         uint16 = 0x02
	LocateTLVTagsInfoUnavailableMime uint16 = 0x03
	LocateTLVTagsInfoUnavailableData uint16 = 0x04
	LocateTLVTagsInfoCapabilities    uint16 = 0x05
	LocateTLVTagsInfoCerts           uint16 = 0x06
	LocateTLVTagsInfoSigTime         uint16 = 0x0A
	LocateTLVTagsInfoUnavailableTime uint16 = 0x0B
	LocateTLVTagsInfoSupportHostSig  uint16 = 0x0C
	LocateTLVTagsInfoHtmlInfoData    uint16 = 0x0E
	LocateTLVTagsInfoHtmlInfoType    uint16 = 0x0D
)

func routeLocate(sess *Session, sm SessionManager, fm *FeedbagStore, snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch snac.SubGroup {
	case LocateRightsQuery:
		return SendAndReceiveLocateRights(snac, w, sequence)
	case LocateSetInfo:
		return ReceiveSetInfo(sess, sm, fm, snac, r)
	case LocateSetDirInfo:
		return SendAndReceiveSetDirInfo(snac, r, w, sequence)
	case LocateGetDirInfo:
		return ReceiveLocateGetDirInfo(snac, r)
	case LocateSetKeywordInfo:
		return SendAndReceiveSetKeywordInfo(snac, r, w, sequence)
	case LocateUserInfoQuery2:
		return SendAndReceiveUserInfoQuery2(sess, sm, fm, fm, snac, r, w, sequence)
	default:
		return handleUnimplementedSNAC(snac, w, sequence)
	}
}

func SendAndReceiveLocateRights(snac oscar.SnacFrame, w io.Writer, sequence *uint32) error {
	fmt.Printf("sendAndReceiveLocateRights read SNAC frame: %+v\n", snac)

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: LOCATE,
		SubGroup:  LocateRightsReply,
	}
	snacPayloadOut := oscar.SNAC_0x02_0x03_LocateRightsReply{
		TLVRestBlock: oscar.TLVRestBlock{
			TLVList: oscar.TLVList{
				{
					TType: 0x01,
					Val:   uint16(1000),
				},
				{
					TType: 0x02,
					Val:   uint16(1000),
				},
				{
					TType: 0x03,
					Val:   uint16(1000),
				},
				{
					TType: 0x04,
					Val:   uint16(1000),
				},
				{
					TType: 0x05,
					Val:   uint16(1000),
				},
			},
		},
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func ReceiveSetInfo(sess *Session, sm SessionManager, fm *FeedbagStore, snac oscar.SnacFrame, r io.Reader) error {
	fmt.Printf("ReceiveSetInfo read SNAC frame: %+v\n", snac)

	snacPayloadIn := oscar.SNAC_0x02_0x04_LocateSetInfo{}
	if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	// update profile
	if profile, hasProfile := snacPayloadIn.GetString(LocateTLVTagsInfoSigData); hasProfile {
		if err := fm.UpsertProfile(sess.ScreenName, profile); err != nil {
			return err
		}
	}

	// broadcast away message change to buddies
	if awayMsg, hasAwayMsg := snacPayloadIn.GetString(LocateTLVTagsInfoUnavailableData); hasAwayMsg {
		sess.SetAwayMessage(awayMsg)
		if err := NotifyArrival(sess, sm, fm); err != nil {
			return err
		}
	}

	fmt.Printf("ReceiveSetInfo read SNAC: %+v\n", snacPayloadIn)

	return nil
}

func ReceiveLocateGetDirInfo(snac oscar.SnacFrame, r io.Reader) error {
	fmt.Printf("ReceiveLocateGetDirInfo read SNAC frame: %+v\n", snac)

	snacPayloadIn := oscar.SNAC_0x02_0x0B_LocateGetDirInfo{}
	if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	fmt.Printf("ReceiveLocateGetDirInfo read SNAC: %+v\n", snacPayloadIn)

	return nil
}

func SendAndReceiveUserInfoQuery2(sess *Session, sm SessionManager, fm FeedbagManager, pm ProfileManager, snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	snacPayloadIn := oscar.SNAC_0x02_0x15_LocateUserInfoQuery2{}
	if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	blocked, err := fm.Blocked(sess.ScreenName, snacPayloadIn.ScreenName)
	if err != nil {
		return err
	}
	if blocked != BlockedNo {
		snacFrameOut := oscar.SnacFrame{
			FoodGroup: LOCATE,
			SubGroup:  LocateErr,
		}
		snacPayloadOut := oscar.SnacError{
			Code: ErrorCodeNotLoggedOn,
		}
		return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
	}

	buddySess, err := sm.RetrieveByScreenName(snacPayloadIn.ScreenName)
	if err != nil {
		if errors.Is(err, errSessNotFound) {
			snacFrameOut := oscar.SnacFrame{
				FoodGroup: LOCATE,
				SubGroup:  LocateErr,
			}
			snacPayloadOut := oscar.SnacError{
				Code: ErrorCodeNotLoggedOn,
			}
			return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
		}
		return err
	}

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: LOCATE,
		SubGroup:  LocateUserInfoReply,
	}
	snacPayloadOut := oscar.SNAC_0x02_0x06_LocateUserInfoReply{
		TLVUserInfo: buddySess.GetTLVUserInfo(),
	}

	if snacPayloadIn.RequestProfile() {
		profile, err := pm.RetrieveProfile(snacPayloadIn.ScreenName)
		if err != nil {
			return err
		}
		snacPayloadOut.LocateInfo.TLVList = append(snacPayloadOut.LocateInfo.TLVList, []oscar.TLV{
			{
				TType: LocateTLVTagsInfoSigMime,
				Val:   `text/aolrtf; charset="us-ascii"`,
			},
			{
				TType: LocateTLVTagsInfoSigData,
				Val:   profile,
			},
		}...)
	}

	if snacPayloadIn.RequestAwayMessage() {
		snacPayloadOut.LocateInfo.TLVList = append(snacPayloadOut.LocateInfo.TLVList, []oscar.TLV{
			{
				TType: LocateTLVTagsInfoUnavailableMime,
				Val:   `text/aolrtf; charset="us-ascii"`,
			},
			{
				TType: LocateTLVTagsInfoUnavailableData,
				Val:   buddySess.GetAwayMessage(),
			},
		}...)
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func SendAndReceiveSetDirInfo(snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("SendAndReceiveSetDirInfo read SNAC frame: %+v\n", snac)

	snacPayloadIn := oscar.SNAC_0x02_0x09_LocateSetDirInfo{}
	if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: LOCATE,
		SubGroup:  LocateSetDirReply,
	}
	snacPayloadOut := oscar.SNAC_0x02_0x0A_LocateSetDirReply{
		Result: 1,
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func SendAndReceiveSetKeywordInfo(snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("SendAndReceiveSetKeywordInfo read SNAC frame: %+v\n", snac)

	snacPayloadIn := oscar.SNAC_0x02_0x0F_LocateSetKeywordInfo{}
	if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: LOCATE,
		SubGroup:  LocateSetKeywordReply,
	}
	snacPayloadOut := oscar.SNAC_0x02_0x10_LocateSetKeywordReply{
		Unknown: 1,
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}
