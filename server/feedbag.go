package server

import (
	"errors"
	"fmt"
	"github.com/mkaminski/goaim/oscar"
	"io"
	"time"
)

const (
	FeedbagErr                      uint16 = 0x0001
	FeedbagRightsQuery                     = 0x0002
	FeedbagQuery                           = 0x0004
	FeedbagQueryIfModified                 = 0x0005
	FeedbagReply                           = 0x0006
	FeedbagUse                             = 0x0007
	FeedbagInsertItem                      = 0x0008
	FeedbagUpdateItem                      = 0x0009
	FeedbagDeleteItem                      = 0x000A
	FeedbagInsertClass                     = 0x000B
	FeedbagUpdateClass                     = 0x000C
	FeedbagDeleteClass                     = 0x000D
	FeedbagStatus                          = 0x000E
	FeedbagReplyNotModified                = 0x000F
	FeedbagDeleteUser                      = 0x0010
	FeedbagStartCluster                    = 0x0011
	FeedbagEndCluster                      = 0x0012
	FeedbagAuthorizeBuddy                  = 0x0013
	FeedbagPreAuthorizeBuddy               = 0x0014
	FeedbagPreAuthorizedBuddy              = 0x0015
	FeedbagRemoveMe                        = 0x0016
	FeedbagRemoveMe2                       = 0x0017
	FeedbagRequestAuthorizeToHost          = 0x0018
	FeedbagRequestAuthorizeToClient        = 0x0019
	FeedbagRespondAuthorizeToHost          = 0x001A
	FeedbagRespondAuthorizeToClient        = 0x001B
	FeedbagBuddyAdded                      = 0x001C
	FeedbagRequestAuthorizeToBadog         = 0x001D
	FeedbagRespondAuthorizeToBadog         = 0x001E
	FeedbagBuddyAddedToBadog               = 0x001F
	FeedbagTestSnac                        = 0x0021
	FeedbagForwardMsg                      = 0x0022
	FeedbagIsAuthRequiredQuery             = 0x0023
	FeedbagIsAuthRequiredReply             = 0x0024
	FeedbagRecentBuddyUpdate               = 0x0025
)

const (
	FeedbagAttributesShared                  uint16 = 0x0064
	FeedbagAttributesInvited                        = 0x0065
	FeedbagAttributesPending                        = 0x0066
	FeedbagAttributesTimeT                          = 0x0067
	FeedbagAttributesDenied                         = 0x0068
	FeedbagAttributesSwimIndex                      = 0x0069
	FeedbagAttributesRecentBuddy                    = 0x006A
	FeedbagAttributesAutoBot                        = 0x006B
	FeedbagAttributesInteraction                    = 0x006D
	FeedbagAttributesMegaBot                        = 0x006F
	FeedbagAttributesOrder                          = 0x00C8
	FeedbagAttributesBuddyPrefs                     = 0x00C9
	FeedbagAttributesPdMode                         = 0x00CA
	FeedbagAttributesPdMask                         = 0x00CB
	FeedbagAttributesPdFlags                        = 0x00CC
	FeedbagAttributesClientPrefs                    = 0x00CD
	FeedbagAttributesLanguage                       = 0x00CE
	FeedbagAttributesFishUri                        = 0x00CF
	FeedbagAttributesWirelessPdMode                 = 0x00D0
	FeedbagAttributesWirelessIgnoreMode             = 0x00D1
	FeedbagAttributesFishPdMode                     = 0x00D2
	FeedbagAttributesFishIgnoreMode                 = 0x00D3
	FeedbagAttributesCreateTime                     = 0x00D4
	FeedbagAttributesBartInfo                       = 0x00D5
	FeedbagAttributesBuddyPrefsValid                = 0x00D6
	FeedbagAttributesBuddyPrefs2                    = 0x00D7
	FeedbagAttributesBuddyPrefs2Valid               = 0x00D8
	FeedbagAttributesBartList                       = 0x00D9
	FeedbagAttributesArriveSound                    = 0x012C
	FeedbagAttributesLeaveSound                     = 0x012D
	FeedbagAttributesImage                          = 0x012E
	FeedbagAttributesColorBg                        = 0x012F
	FeedbagAttributesColorFg                        = 0x0130
	FeedbagAttributesAlias                          = 0x0131
	FeedbagAttributesPassword                       = 0x0132
	FeedbagAttributesDisabled                       = 0x0133
	FeedbagAttributesCollapsed                      = 0x0134
	FeedbagAttributesUrl                            = 0x0135
	FeedbagAttributesActiveList                     = 0x0136
	FeedbagAttributesEmailAddr                      = 0x0137
	FeedbagAttributesPhoneNumber                    = 0x0138
	FeedbagAttributesCellPhoneNumber                = 0x0139
	FeedbagAttributesSmsPhoneNumber                 = 0x013A
	FeedbagAttributesWireless                       = 0x013B
	FeedbagAttributesNote                           = 0x013C
	FeedbagAttributesAlertPrefs                     = 0x013D
	FeedbagAttributesBudalertSound                  = 0x013E
	FeedbagAttributesStockalertValue                = 0x013F
	FeedbagAttributesTpalertEditUrl                 = 0x0140
	FeedbagAttributesTpalertDeleteUrl               = 0x0141
	FeedbagAttributesTpprovMorealertsUrl            = 0x0142
	FeedbagAttributesFish                           = 0x0143
	FeedbagAttributesXunconfirmedxLastAccess        = 0x0145
	FeedbagAttributesImSent                         = 0x0150
	FeedbagAttributesOnlineTime                     = 0x0151
	FeedbagAttributesAwayMsg                        = 0x0152
	FeedbagAttributesImReceived                     = 0x0153
	FeedbagAttributesBuddyfeedView                  = 0x0154
	FeedbagAttributesWorkPhoneNumber                = 0x0158
	FeedbagAttributesOtherPhoneNumber               = 0x0159
	FeedbagAttributesWebPdMode                      = 0x015F
	FeedbagAttributesFirstCreationTimeXc            = 0x0167
	FeedbagAttributesPdModeXc                       = 0x016E
)
const (
	FeedbagClassIdBuddy            uint16 = 0x0000
	FeedbagClassIdGroup                   = 0x0001
	FeedbagClassIdPermit                  = 0x0002
	FeedbagClassIdDeny                    = 0x0003
	FeedbagClassIdPdinfo                  = 0x0004
	FeedbagClassIdBuddyPrefs              = 0x0005
	FeedbagClassIdNonbuddy                = 0x0006
	FeedbagClassIdTpaProvider             = 0x0007
	FeedbagClassIdTpaSubscription         = 0x0008
	FeedbagClassIdClientPrefs             = 0x0009
	FeedbagClassIdStock                   = 0x000A
	FeedbagClassIdWeather                 = 0x000B
	FeedbagClassIdWatchList               = 0x000D
	FeedbagClassIdIgnoreList              = 0x000E
	FeedbagClassIdDateTime                = 0x000F
	FeedbagClassIdExternalUser            = 0x0010
	FeedbagClassIdRootCreator             = 0x0011
	FeedbagClassIdFish                    = 0x0012
	FeedbagClassIdImportTimestamp         = 0x0013
	FeedbagClassIdBart                    = 0x0014
	FeedbagClassIdRbOrder                 = 0x0015
	FeedbagClassIdPersonality             = 0x0016
	FeedbagClassIdAlProf                  = 0x0017
	FeedbagClassIdAlInfo                  = 0x0018
	FeedbagClassIdInteraction             = 0x0019
	FeedbagClassIdVanityInfo              = 0x001D
	FeedbagClassIdFavoriteLocation        = 0x001E
	FeedbagClassIdBartPdinfo              = 0x001F
	FeedbagClassIdCustomEmoticons         = 0x0024
	FeedbagClassIdMaxPredefined           = 0x0024
	FeedbagClassIdXIcqStatusNote          = 0x015C
	FeedbagClassIdMin                     = 0x0400
)

func routeFeedbag(sm SessionManager, sess *Session, fm *FeedbagStore, snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch snac.SubGroup {
	case FeedbagErr:
		panic("not implemented")
	case FeedbagRightsQuery:
		return SendAndReceiveFeedbagRightsQuery(snac, r, w, sequence)
	case FeedbagQuery:
		return ReceiveAndSendFeedbagQuery(sess, fm, snac, w, sequence)
	case FeedbagQueryIfModified:
		return ReceiveAndSendFeedbagQueryIfModified(sess, fm, snac, r, w, sequence)
	case FeedbagUse:
		return ReceiveUse(snac, r, w, sequence)
	case FeedbagInsertItem:
		return ReceiveInsertItem(sm, sess, fm, snac, r, w, sequence)
	case FeedbagUpdateItem:
		return ReceiveUpdateItem(sm, sess, fm, snac, r, w, sequence)
	case FeedbagDeleteItem:
		return ReceiveDeleteItem(sm, sess, fm, snac, r, w, sequence)
	case FeedbagInsertClass:
		panic("not implemented")
	case FeedbagUpdateClass:
		panic("not implemented")
	case FeedbagDeleteClass:
		panic("not implemented")
	case FeedbagStatus:
		panic("not implemented")
	case FeedbagReplyNotModified:
		panic("not implemented")
	case FeedbagDeleteUser:
		panic("not implemented")
	case FeedbagStartCluster:
		return ReceiveFeedbagStartCluster(snac, r, w, sequence)
	case FeedbagEndCluster:
		return ReceiveFeedbagEndCluster(snac, r, w, sequence)
	case FeedbagAuthorizeBuddy:
		panic("not implemented")
	case FeedbagPreAuthorizeBuddy:
		panic("not implemented")
	case FeedbagPreAuthorizedBuddy:
		panic("not implemented")
	case FeedbagRemoveMe:
		panic("not implemented")
	case FeedbagRemoveMe2:
		panic("not implemented")
	case FeedbagRequestAuthorizeToHost:
		panic("not implemented")
	case FeedbagRequestAuthorizeToClient:
		panic("not implemented")
	case FeedbagRespondAuthorizeToHost:
		panic("not implemented")
	case FeedbagRespondAuthorizeToClient:
		panic("not implemented")
	case FeedbagBuddyAdded:
		panic("not implemented")
	case FeedbagRequestAuthorizeToBadog:
		panic("not implemented")
	case FeedbagRespondAuthorizeToBadog:
		panic("not implemented")
	case FeedbagBuddyAddedToBadog:
		panic("not implemented")
	case FeedbagTestSnac:
		panic("not implemented")
	case FeedbagForwardMsg:
		panic("not implemented")
	case FeedbagIsAuthRequiredQuery:
		panic("not implemented")
	case FeedbagIsAuthRequiredReply:
		panic("not implemented")
	case FeedbagRecentBuddyUpdate:
		panic("not implemented")
	}

	return nil
}

func SendAndReceiveFeedbagRightsQuery(snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("sendAndReceiveFeedbagRightsQuery read SNAC frame: %+v\n", snac)

	snacPayloadIn := oscar.SNAC_0x13_0x02_FeedbagRightsQuery{}
	if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	fmt.Printf("sendAndReceiveFeedbagRightsQuery read SNAC payload: %+v\n", snacPayloadIn)

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: 0x13,
		SubGroup:  0x03,
	}
	snacPayloadOut := oscar.SNAC_0x13_0x03_FeedbagRightsReply{
		TLVRestBlock: oscar.TLVRestBlock{
			TLVList: oscar.TLVList{
				{
					TType: 0x03,
					Val:   uint16(200),
				},
				{
					TType: 0x04,
					Val: []uint16{
						0x3D,
						0x3D,
						0x64,
						0x64,
						0x01,
						0x01,
						0x32,
						0x00,
						0x00,
						0x03,
						0x00,
						0x00,
						0x00,
						0x80,
						0xFF,
						0x14,
						0xC8,
						0x01,
						0x00,
						0x01,
						0x00,
					},
				},
				{
					TType: 0x05,
					Val:   uint16(200),
				},
				{
					TType: 0x06,
					Val:   uint16(200),
				},
				{
					TType: 0x07,
					Val:   uint16(200),
				},
				{
					TType: 0x08,
					Val:   uint16(200),
				},
				{
					TType: 0x09,
					Val:   uint16(200),
				},
				{
					TType: 0x0A,
					Val:   uint16(200),
				},
				{
					TType: 0x0C,
					Val:   uint16(200),
				},
				{
					TType: 0x0D,
					Val:   uint16(200),
				},
				{
					TType: 0x0E,
					Val:   uint16(100),
				},
			},
		},
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func ReceiveAndSendFeedbagQuery(sess *Session, fm FeedbagManager, snac oscar.SnacFrame, w io.Writer, sequence *uint32) error {
	fb, err := fm.Retrieve(sess.ScreenName)
	if err != nil {
		return err
	}

	lm := time.UnixMilli(0)

	if len(fb) > 0 {
		lm, err = fm.LastModified(sess.ScreenName)
		if err != nil {
			return err
		}
	}

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: FEEDBAG,
		SubGroup:  FeedbagReply,
	}
	snacPayloadOut := oscar.SNAC_0x13_0x06_FeedbagReply{
		Version:    0,
		Items:      fb,
		LastUpdate: uint32(lm.Unix()),
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func ReceiveAndSendFeedbagQueryIfModified(sess *Session, fm FeedbagManager, snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	snacPayloadIn := oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{}
	if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	fb, err := fm.Retrieve(sess.ScreenName)
	if err != nil {
		return err
	}

	lm := time.UnixMilli(0)

	if len(fb) > 0 {
		lm, err = fm.LastModified(sess.ScreenName)
		if err != nil {
			return err
		}
		if lm.Before(time.Unix(int64(snacPayloadIn.LastUpdate), 0)) {
			snacFrameOut := oscar.SnacFrame{
				FoodGroup: FEEDBAG,
				SubGroup:  FeedbagReplyNotModified,
			}
			snacPayloadOut := oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
				LastUpdate: uint32(lm.Unix()),
				Count:      uint8(len(fb)),
			}
			return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
		}
	}

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: FEEDBAG,
		SubGroup:  FeedbagReply,
	}
	snacPayloadOut := oscar.SNAC_0x13_0x06_FeedbagReply{
		Version:    0,
		Items:      fb,
		LastUpdate: uint32(lm.Unix()),
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func ReceiveInsertItem(sm SessionManager, sess *Session, fm FeedbagManager, snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	snacPayloadIn := oscar.SNAC_0x13_0x08_FeedbagInsertItem{}
	if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	for _, item := range snacPayloadIn.Items {
		// don't let users block themselves, it causes the AIM client to go
		// into a weird state.
		if item.ClassID == 3 && item.Name == sess.ScreenName {
			snacFrameOut := oscar.SnacFrame{
				FoodGroup: FEEDBAG,
				SubGroup:  FeedbagErr,
			}
			snacPayloadOut := oscar.SnacError{
				Code: ErrorCodeNotSupportedByHost,
			}
			return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
		}
	}

	if err := fm.Upsert(sess.ScreenName, snacPayloadIn.Items); err != nil {
		return err
	}

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: FEEDBAG,
		SubGroup:  FeedbagStatus,
	}
	snacPayloadOut := oscar.SNAC_0x13_0x0E_FeedbagStatus{}

	for range snacPayloadIn.Items {
		snacPayloadOut.Results = append(snacPayloadOut.Results, 0x0000)
	}

	if err := writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w); err != nil {
		return err
	}

	for _, item := range snacPayloadIn.Items {
		switch item.ClassID {
		case 2:
			// notify that added buddy is online
			if err := NotifyBuddyOnline(w, item.Name, sm, sequence); err != nil {
				return err
			}
		case 3:
			// DENY, block buddy
			if err := blockBuddy(sm, sess, item.Name, sequence, w); err != nil {
				return err
			}
		}
	}
	return nil
}

func blockBuddy(sm SessionManager, sess *Session, screenName string, sequence *uint32, w io.Writer) error {
	buddySess, err := sm.RetrieveByScreenName(screenName)
	if err != nil {
		if errors.Is(err, errSessNotFound) {
			// former buddy is offline
			return nil
		}
		return err
	}

	// tell the blocked buddy you've signed off
	sm.SendToScreenName(screenName, XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: BUDDY,
			SubGroup:  BuddyDeparted,
		},
		snacOut: oscar.SNAC_0x03_0x0A_BuddyArrived{
			TLVUserInfo: oscar.TLVUserInfo{
				ScreenName:   sess.ScreenName,
				WarningLevel: sess.GetWarning(),
			},
		},
	})

	// tell yourself blocked buddy has signed off
	snacFrameOut := oscar.SnacFrame{
		FoodGroup: BUDDY,
		SubGroup:  BuddyDeparted,
	}
	snacPayloadOut := oscar.SNAC_0x03_0x0A_BuddyArrived{
		TLVUserInfo: oscar.TLVUserInfo{
			ScreenName:   buddySess.ScreenName,
			WarningLevel: buddySess.GetWarning(),
		},
	}

	return writeOutSNAC(oscar.SnacFrame{}, snacFrameOut, snacPayloadOut, sequence, w)
}

func ReceiveUpdateItem(sm SessionManager, sess *Session, fm *FeedbagStore, snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("ReceiveUpdateItem read SNAC frame: %+v\n", snac)

	snacPayloadIn := oscar.SNAC_0x13_0x09_FeedbagUpdateItem{}
	if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	if err := fm.Upsert(sess.ScreenName, snacPayloadIn.Items); err != nil {
		return err
	}

	snacPayloadOut := oscar.SNAC_0x13_0x0E_FeedbagStatus{}

	for _, item := range snacPayloadIn.Items {
		snacPayloadOut.Results = append(snacPayloadOut.Results, 0x0000) // success by default
		fmt.Printf("ReceiveUpdateItem read SNAC feedbag item: %+v\n", item)
	}

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: FEEDBAG,
		SubGroup:  FeedbagStatus,
	}

	if err := writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w); err != nil {
		return err
	}

	return GetAllOnlineBuddies(w, sess, sm, fm, sequence)
}

func ReceiveDeleteItem(sm SessionManager, sess *Session, fm *FeedbagStore, snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("ReceiveUpdateItem read SNAC frame: %+v\n", snac)

	snacPayloadIn := oscar.SNAC_0x13_0x0A_FeedbagDeleteItem{}
	if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	if err := fm.Delete(sess.ScreenName, snacPayloadIn.Items); err != nil {
		return err
	}

	snacPayloadOut := oscar.SNAC_0x13_0x0E_FeedbagStatus{}

	hasUnblock := false
	for _, item := range snacPayloadIn.Items {
		if item.ClassID == 3 {
			hasUnblock = true
		}
		snacPayloadOut.Results = append(snacPayloadOut.Results, 0x0000) // success by default
		fmt.Printf("ReceiveDeleteItem read SNAC feedbag item: %+v\n", item)
	}

	if hasUnblock {
		// notify previously blocked users that user is back online
		if err := NotifyArrival(sess, sm, fm); err != nil {
			return err
		}
	}

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: FEEDBAG,
		SubGroup:  FeedbagStatus,
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func ReceiveFeedbagStartCluster(snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("ReceiveFeedbagStartCluster read SNAC frame: %+v\n", snac)
	tlv := oscar.TLVRestBlock{}
	return tlv.Read(r)
}

func ReceiveFeedbagEndCluster(snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("receiveFeedbagEndCluster read SNAC frame: %+v\n", snac)
	return nil
}

func ReceiveUse(snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("ReceiveUse read SNAC frame: %+v\n", snac)
	return nil
}
