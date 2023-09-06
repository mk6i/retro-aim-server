package oscar

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	FeedbagErr                      uint16 = 0x0001
	FeedbagRightsQuery                     = 0x0002
	FeedbagQuery                           = 0x0004
	FeedbagQueryIfModified                 = 0x0005
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

func routeFeedbag(sm *SessionManager, sess *Session, fm *FeedbagStore, snac snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch snac.subGroup {
	case FeedbagErr:
		panic("not implemented")
	case FeedbagRightsQuery:
		return SendAndReceiveFeedbagRightsQuery(snac, r, w, sequence)
	case FeedbagQuery:
		return ReceiveAndSendFeedbagQuery(sess, fm, snac, r, w, sequence)
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

type payloadFeedbagRightsQuery struct {
	TLVPayload
}

func (s *payloadFeedbagRightsQuery) read(r io.Reader) error {
	return s.TLVPayload.read(r)
}

func SendAndReceiveFeedbagRightsQuery(snac snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("sendAndReceiveFeedbagRightsQuery read SNAC frame: %+v\n", snac)

	snacPayloadIn := payloadFeedbagRightsQuery{}
	if err := snacPayloadIn.read(r); err != nil {
		return err
	}

	fmt.Printf("sendAndReceiveFeedbagRightsQuery read SNAC payload: %+v\n", snacPayloadIn)

	snacFrameOut := snacFrame{
		foodGroup: 0x13,
		subGroup:  0x03,
	}
	snacPayloadOut := TLVPayload{
		TLVs: []TLV{
			{
				tType: 0x03,
				val:   uint16(200),
			},
			{
				tType: 0x04,
				val: []uint16{
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
				tType: 0x05,
				val:   uint16(200),
			},
			{
				tType: 0x06,
				val:   uint16(200),
			},
			{
				tType: 0x07,
				val:   uint16(200),
			},
			{
				tType: 0x08,
				val:   uint16(200),
			},
			{
				tType: 0x09,
				val:   uint16(200),
			},
			{
				tType: 0x0A,
				val:   uint16(200),
			},
			{
				tType: 0x0C,
				val:   uint16(200),
			},
			{
				tType: 0x0D,
				val:   uint16(200),
			},
			{
				tType: 0x0E,
				val:   uint16(100),
			},
		},
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

type feedbagItem struct {
	name    string
	groupID uint16
	itemID  uint16
	classID uint16
	TLVPayload
}

func (f feedbagItem) write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, uint16(len(f.name))); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, []byte(f.name)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, f.groupID); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, f.itemID); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, f.classID); err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	if err := f.TLVPayload.write(buf); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint16(buf.Len())); err != nil {
		return err
	}
	_, err := w.Write(buf.Bytes())
	return err
}

func (f *feedbagItem) read(r io.Reader) error {
	var l uint16
	if err := binary.Read(r, binary.BigEndian, &l); err != nil {
		return err
	}
	buf := make([]byte, l)
	if _, err := r.Read(buf); err != nil {
		return err
	}
	f.name = string(buf)
	if err := binary.Read(r, binary.BigEndian, &f.groupID); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &f.itemID); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &f.classID); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &l); err != nil {
		return err
	}
	buf = make([]byte, l)
	if _, err := r.Read(buf); err != nil {
		return err
	}

	return f.TLVPayload.read(bytes.NewBuffer(buf))
}

type snacFeedbagQuery struct {
	version    uint8
	items      []feedbagItem
	lastUpdate uint32
}

func (s snacFeedbagQuery) write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, s.version); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint16(len(s.items))); err != nil {
		return err
	}
	for _, t := range s.items {
		if err := t.write(w); err != nil {
			return err
		}
	}
	return binary.Write(w, binary.BigEndian, s.lastUpdate)
}

func ReceiveAndSendFeedbagQuery(sess *Session, fm *FeedbagStore, snac snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("receiveAndSendFeedbagQuery read SNAC frame: %+v\n", snac)

	fb, err := fm.Retrieve(sess.ScreenName)
	if err != nil {
		return err
	}

	var lastModified uint32
	if len(fb) > 0 {
		lm, err := fm.LastModified(sess.ScreenName)
		if err != nil {
			return err
		}
		lastModified = uint32(lm.Unix())
	}

	snacFrameOut := snacFrame{
		foodGroup: 0x13,
		subGroup:  0x06,
	}
	snacPayloadOut := snacFeedbagQuery{
		version:    0,
		items:      fb,
		lastUpdate: lastModified,
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

type snacQueryIfModified struct {
	lastUpdate uint32
	count      uint8
}

func (s snacQueryIfModified) write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, s.lastUpdate); err != nil {
		return err
	}
	return binary.Write(w, binary.BigEndian, s.count)
}

func (s *snacQueryIfModified) read(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &s.lastUpdate); err != nil {
		return err
	}
	return binary.Read(r, binary.BigEndian, &s.count)
}

func ReceiveAndSendFeedbagQueryIfModified(sess *Session, fm *FeedbagStore, snac snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("ReceiveAndSendFeedbagQueryIfModified read SNAC frame: %+v\n", snac)

	snacPayload := snacQueryIfModified{}
	if err := snacPayload.read(r); err != nil && err != io.EOF {
		return err
	}

	fmt.Printf("ReceiveAndSendFeedbagQueryIfModified read SNAC: %+v\n", snacPayload)

	fb, err := fm.Retrieve(sess.ScreenName)
	if err != nil {
		return err
	}

	lm, err := fm.LastModified(sess.ScreenName)
	if err != nil {
		return err
	}

	//if lm.Before(time.Unix(int64(snacPayload.lastUpdate), 0)) {
	//todo not sure this works right now
	//	snacFrameOut := snacFrame{
	//		foodGroup: 0x13,
	//		subGroup:  0x0F,
	//	}
	//	lm, err := fm.LastModified(sess.ScreenName)
	//	if err != nil {
	//		return err
	//	}
	//	snacPayloadOut := snacQueryIfModified{
	//		lastUpdate: uint32(lm.Unix()),
	//		count:      uint8(len(fb)),
	//	}
	//	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
	//}

	snacFrameOut := snacFrame{
		foodGroup: 0x13,
		subGroup:  0x06,
	}
	snacPayloadOut := snacFeedbagQuery{
		version:    0,
		items:      fb,
		lastUpdate: uint32(lm.Unix()),
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

type snacFeedbagStatusReply struct {
	results []uint16
}

func (s snacFeedbagStatusReply) write(w io.Writer) error {
	return binary.Write(w, binary.BigEndian, s.results)
}

func ReceiveInsertItem(sm *SessionManager, sess *Session, fm *FeedbagStore, snac snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("ReceiveInsertItem read SNAC frame: %+v\n", snac)

	snacPayloadOut := snacFeedbagStatusReply{}
	var feedbag []feedbagItem

	for {
		item := feedbagItem{}
		if err := item.read(r); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		// don't let users block themselves, it causes the AIM client to go
		// into a weird state.
		if item.classID == 3 && item.name == sess.ScreenName {
			snacFrameOut := snacFrame{
				foodGroup: FEEDBAG,
				subGroup:  FeedbagErr,
			}
			snacPayloadOut := snacError{
				code: ErrorCodeNotSupportedByHost,
			}
			return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
		}
		feedbag = append(feedbag, item)
		snacPayloadOut.results = append(snacPayloadOut.results, 0x0000) // success by default
		fmt.Printf("ReceiveInsertItem read SNAC feedbag item: %+v\n", item)
	}

	if err := fm.Upsert(sess.ScreenName, feedbag); err != nil {
		return err
	}

	snacFrameOut := snacFrame{
		foodGroup: FEEDBAG,
		subGroup:  FeedbagStatus,
	}

	if err := writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w); err != nil {
		return err
	}

	for _, item := range feedbag {
		// DENY, block buddy
		if item.classID == 3 {
			if err := blockBuddy(sm, sess, item.name, sequence, w); err != nil {
				return err
			}
		}
	}

	// todo: just check online status for buddies that were added
	return GetOnlineBuddies(w, sess, sm, fm, sequence)
}

func blockBuddy(sm *SessionManager, sess *Session, screenName string, sequence *uint32, w io.Writer) error {
	// tell the blocked buddy you've signed off
	sm.SendToScreenName(screenName, XMessage{
		snacFrame: snacFrame{
			foodGroup: BUDDY,
			subGroup:  BuddyDeparted,
		},
		snacOut: snacBuddyArrived{
			screenName:   sess.ScreenName,
			warningLevel: sess.GetWarning(),
		},
	})

	// show your blocked buddy as signed off
	buddySess, err := sm.RetrieveByScreenName(screenName)
	if err != nil {
		if errors.Is(err, errSessNotFound) {
			// former buddy is offline
			return nil
		}
		return err
	}

	snacFrameOut := snacFrame{
		foodGroup: BUDDY,
		subGroup:  BuddyDeparted,
	}
	snacPayloadOut := snacBuddyArrived{
		screenName:   buddySess.ScreenName,
		warningLevel: buddySess.GetWarning(),
	}

	return writeOutSNAC(snacFrame{}, snacFrameOut, snacPayloadOut, sequence, w)
}

func ReceiveUpdateItem(sm *SessionManager, sess *Session, fm *FeedbagStore, snac snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("ReceiveUpdateItem read SNAC frame: %+v\n", snac)

	var items []feedbagItem

	for {
		item := feedbagItem{}
		if err := item.read(r); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		fmt.Printf("\titem: %+v\n", item)
		items = append(items, item)
	}

	if err := fm.Upsert(sess.ScreenName, items); err != nil {
		return err
	}

	snacPayloadOut := snacFeedbagStatusReply{}

	for _, item := range items {
		snacPayloadOut.results = append(snacPayloadOut.results, 0x0000) // success by default
		fmt.Printf("ReceiveUpdateItem read SNAC feedbag item: %+v\n", item)
	}

	snacFrameOut := snacFrame{
		foodGroup: FEEDBAG,
		subGroup:  FeedbagStatus,
	}

	if err := writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w); err != nil {
		return err
	}

	return GetOnlineBuddies(w, sess, sm, fm, sequence)
}

func ReceiveDeleteItem(sm *SessionManager, sess *Session, fm *FeedbagStore, snac snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("ReceiveUpdateItem read SNAC frame: %+v\n", snac)

	var items []feedbagItem

	for {
		item := feedbagItem{}
		if err := item.read(r); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		fmt.Printf("\titem: %+v\n", item)
		items = append(items, item)
	}

	if err := fm.Delete(sess.ScreenName, items); err != nil {
		return err
	}

	snacPayloadOut := snacFeedbagStatusReply{}

	hasUnblock := false
	for _, item := range items {
		if item.classID == 3 {
			hasUnblock = true
		}
		snacPayloadOut.results = append(snacPayloadOut.results, 0x0000) // success by default
		fmt.Printf("ReceiveDeleteItem read SNAC feedbag item: %+v\n", item)
	}

	if hasUnblock {
		// notify previously blocked users that user is back online
		if err := NotifyArrival(sess, sm, fm); err != nil {
			return err
		}
	}

	snacFrameOut := snacFrame{
		foodGroup: FEEDBAG,
		subGroup:  FeedbagStatus,
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func ReceiveFeedbagStartCluster(snac snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("ReceiveFeedbagStartCluster read SNAC frame: %+v\n", snac)

	tlv := TLVPayload{}
	if err := tlv.read(r); err != nil {
		return err
	}

	return nil
}

func ReceiveFeedbagEndCluster(snac snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("receiveFeedbagEndCluster read SNAC frame: %+v\n", snac)
	return nil
}

func ReceiveUse(snac snacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("ReceiveUse read SNAC frame: %+v\n", snac)
	return nil
}
