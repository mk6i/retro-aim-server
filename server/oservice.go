package server

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/mkaminski/goaim/oscar"
	"io"
	"time"
)

const (
	OServiceErr               uint16 = 0x0001
	OServiceClientOnline             = 0x0002
	OServiceHostOnline               = 0x0003
	OServiceServiceRequest           = 0x0004
	OServiceServiceResponse          = 0x0005
	OServiceRateParamsQuery          = 0x0006
	OServiceRateParamsReply          = 0x0007
	OServiceRateParamsSubAdd         = 0x0008
	OServiceRateDelParamSub          = 0x0009
	OServiceRateParamChange          = 0x000A
	OServicePauseReq                 = 0x000B
	OServicePauseAck                 = 0x000C
	OServiceResume                   = 0x000D
	OServiceUserInfoQuery            = 0x000E
	OServiceUserInfoUpdate           = 0x000F
	OServiceEvilNotification         = 0x0010
	OServiceIdleNotification         = 0x0011
	OServiceMigrateGroups            = 0x0012
	OServiceMotd                     = 0x0013
	OServiceSetPrivacyFlags          = 0x0014
	OServiceWellKnownUrls            = 0x0015
	OServiceNoop                     = 0x0016
	OServiceClientVersions           = 0x0017
	OServiceHostVersions             = 0x0018
	OServiceMaxConfigQuery           = 0x0019
	OServiceMaxConfigReply           = 0x001A
	OServiceStoreConfig              = 0x001B
	OServiceConfigQuery              = 0x001C
	OServiceConfigReply              = 0x001D
	OServiceSetUserinfoFields        = 0x001E
	OServiceProbeReq                 = 0x001F
	OServiceProbeAck                 = 0x0020
	OServiceBartReply                = 0x0021
	OServiceBartQuery2               = 0x0022
	OServiceBartReply2               = 0x0023
)

func routeOService(cfg Config, ready OnReadyCB, cr *ChatRegistry, sm SessionManager, fm *FeedbagStore, sess *Session, snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch snac.SubGroup {
	case OServiceClientOnline:
		return ReceiveClientOnline(ready, sess, sm, snac, r, w, sequence)
	case OServiceServiceRequest:
		return ReceiveAndSendServiceRequest(cfg, cr, sess, snac, r, w, sequence)
	case OServiceRateParamsQuery:
		return ReceiveAndSendServiceRateParams(snac, r, w, sequence)
	case OServiceRateParamsSubAdd:
		return ReceiveRateParamsSubAdd(snac, r)
	case OServiceUserInfoQuery:
		return ReceiveAndSendServiceRequestSelfInfo(sess, snac, r, w, sequence)
	case OServiceIdleNotification:
		return ReceiveIdleNotification(sess, sm, fm, snac, r)
	case OServiceClientVersions:
		return ReceiveAndSendHostVersions(snac, r, w, sequence)
	case OServiceSetUserinfoFields:
		return ReceiveSetUserInfoFields(sess, sm, fm, snac, r, w, sequence)
	default:
		return ErrUnsupportedSubGroup
	}
}

func WriteOServiceHostOnline(foodGroups []uint16, w io.Writer, sequence *uint32) error {
	fmt.Println("writeOServiceHostOnline...")
	snacFrameOut := oscar.SnacFrame{
		FoodGroup: OSERVICE,
		SubGroup:  OServiceHostOnline,
	}
	snacPayloadOut := oscar.SNAC_0x01_0x03_OServiceHostOnline{
		FoodGroups: foodGroups,
	}
	return writeOutSNAC(oscar.SnacFrame{}, snacFrameOut, snacPayloadOut, sequence, w)
}

func ReceiveAndSendHostVersions(snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("receiveAndSendHostVersions read SNAC frame: %+v\n", snac)

	snacPayloadIn := oscar.SNAC_0x01_0x17_OServiceClientVersions{}
	if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	fmt.Printf("receiveAndSendHostVersions read SNAC: %+v\n", snacPayloadIn)

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: OSERVICE,
		SubGroup:  OServiceHostVersions,
	}
	snacPayloadOut := oscar.SNAC_0x01_0x18_OServiceHostVersions{
		Versions: snacPayloadIn.Versions,
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func ReceiveAndSendServiceRateParams(snac oscar.SnacFrame, _ io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("receiveAndSendServiceRateParams read SNAC frame: %+v\n", snac)

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: OSERVICE,
		SubGroup:  OServiceRateParamsReply,
	}

	snacPayloadOut := oscar.SNAC_0x01_0x07_OServiceRateParamsReply{
		RateClasses: []struct {
			ID              uint16
			WindowSize      uint32
			ClearLevel      uint32
			AlertLevel      uint32
			LimitLevel      uint32
			DisconnectLevel uint32
			CurrentLevel    uint32
			MaxLevel        uint32
			LastTime        uint32 // protocol v2 only
			CurrentState    uint8  // protocol v2 only
		}{
			{
				ID:              0x0001,
				WindowSize:      0x00000050,
				ClearLevel:      0x000009C4,
				AlertLevel:      0x000007D0,
				LimitLevel:      0x000005DC,
				DisconnectLevel: 0x00000320,
				CurrentLevel:    0x00000D69,
				MaxLevel:        0x00001770,
				LastTime:        0x00000000,
				CurrentState:    0x00,
			},
		},
		RateGroups: []struct {
			ID    uint16
			Pairs []struct {
				FoodGroup uint16
				SubGroup  uint16
			} `count_prefix:"uint16"`
		}{
			{
				ID: 1,
				Pairs: []struct {
					FoodGroup uint16
					SubGroup  uint16
				}{},
			},
		},
	}

	for i := uint16(0); i < 24; i++ { // for each food group
		for j := uint16(0); j < 32; j++ { // for each subgroup
			snacPayloadOut.RateGroups[0].Pairs = append(snacPayloadOut.RateGroups[0].Pairs,
				struct {
					FoodGroup uint16
					SubGroup  uint16
				}{
					FoodGroup: i,
					SubGroup:  j,
				})
		}
	}
	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func ReceiveAndSendServiceRequestSelfInfo(sess *Session, snac oscar.SnacFrame, _ io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("receiveAndSendServiceRequestSelfInfo read SNAC frame: %+v\n", snac)

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: OSERVICE,
		SubGroup:  OServiceUserInfoUpdate,
	}
	snacPayloadOut := oscar.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
		TLVUserInfo: oscar.TLVUserInfo{
			ScreenName:   sess.ScreenName,
			WarningLevel: sess.GetWarning(),
			TLVBlock: oscar.TLVBlock{
				TLVList: sess.GetUserInfo(),
			},
		},
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func ReceiveRateParamsSubAdd(snac oscar.SnacFrame, r io.Reader) error {
	fmt.Printf("receiveRateParamsSubAdd read SNAC frame: %+v\n", snac)

	snacPayloadIn := oscar.SNAC_0x01_0x08_OServiceRateParamsSubAdd{}
	if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	fmt.Printf("receiveAndSendHostVersions read SNAC: %+v\n", snacPayloadIn)

	return nil
}

type OnReadyCB func(sess *Session, sm SessionManager, r io.Reader, w io.Writer, sequence *uint32) error

func ReceiveClientOnline(onReadyCB OnReadyCB, sess *Session, sm SessionManager, snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("receiveClientOnline read SNAC frame: %+v\n", snac)

	snacPayloadIn := oscar.SNAC_0x01_0x02_OServiceClientOnline{}
	if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	for _, version := range snacPayloadIn.GroupVersions {
		fmt.Printf("ReceiveClientOnline read SNAC client messageType: %+v\n", version)
	}

	return onReadyCB(sess, sm, r, w, sequence)
}

func GetAllOnlineBuddies(w io.Writer, sess *Session, sm SessionManager, fm FeedbagManager, sequence *uint32) error {
	screenNames, err := fm.Buddies(sess.ScreenName)
	if err != nil {
		return err
	}
	for _, screenName := range screenNames {
		if err := NotifyBuddyOnline(w, screenName, sm, sequence); err != nil {
			return err
		}
	}
	return nil
}

func NotifyBuddyOnline(w io.Writer, screenName string, sm SessionManager, sequence *uint32) error {
	sess, err := sm.RetrieveByScreenName(screenName)
	if err != nil {
		if errors.Is(err, errSessNotFound) {
			// buddy isn't online
			return nil
		}
		return err
	}
	if sess.Invisible() {
		// don't tell user this buddy is online
		return nil
	}

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: BUDDY,
		SubGroup:  BuddyArrived,
	}
	snacPayloadOut := oscar.SNAC_0x03_0x0A_BuddyArrived{
		TLVUserInfo: oscar.TLVUserInfo{
			ScreenName:   screenName,
			WarningLevel: sess.GetWarning(),
			TLVBlock: oscar.TLVBlock{
				TLVList: sess.GetUserInfo(),
			},
		},
	}

	return writeOutSNAC(oscar.SnacFrame{}, snacFrameOut, snacPayloadOut, sequence, w)
}

func ReceiveSetUserInfoFields(sess *Session, sm SessionManager, fm *FeedbagStore, snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	fmt.Printf("receiveSetUserInfoFields read SNAC frame: %+v\n", snac)

	snacPayloadIn := oscar.SNAC_0x01_0x1E_OServiceSetUserInfoFields{}
	if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	if status, hasStatus := snacPayloadIn.GetUint32(0x06); hasStatus {
		switch status {
		case 0x000:
			sess.SetInvisible(false)
			if err := NotifyArrival(sess, sm, fm); err != nil {
				return err
			}
		case 0x100:
			sess.SetInvisible(true)
			if err := NotifyDeparture(sess, sm, fm); err != nil {
				return err
			}
		default:
			return fmt.Errorf("don't know what to do with status %d", status)
		}
	}

	snacFrameOut := oscar.SnacFrame{
		FoodGroup: OSERVICE,
		SubGroup:  OServiceUserInfoUpdate,
	}
	snacPayloadOut := oscar.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
		TLVUserInfo: oscar.TLVUserInfo{
			ScreenName:   sess.ScreenName,
			WarningLevel: sess.GetWarning(),
			TLVBlock: oscar.TLVBlock{
				TLVList: sess.GetUserInfo(),
			},
		},
	}

	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}

func ReceiveIdleNotification(sess *Session, sm SessionManager, fm *FeedbagStore, snac oscar.SnacFrame, r io.Reader) error {
	fmt.Printf("receiveIdleNotification read SNAC frame: %+v\n", snac)

	snacPayloadIn := oscar.SNAC_0x01_0x11_OServiceIdleNotification{}
	if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	if snacPayloadIn.IdleTime == 0 {
		sess.SetActive()
	} else {
		sess.SetIdle(time.Duration(snacPayloadIn.IdleTime) * time.Second)
	}

	return NotifyArrival(sess, sm, fm)
}

const (
	OserviceTlvTagsReconnectHere uint16 = 0x05
	OserviceTlvTagsLoginCookie          = 0x06
	OserviceTlvTagsGroupId              = 0x0D
	OserviceTlvTagsSslCertname          = 0x8D
	OserviceTlvTagsSslState             = 0x8E
)

func ReceiveAndSendServiceRequest(cfg Config, cr *ChatRegistry, sess *Session, snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	snacPayloadIn := oscar.SNAC_0x01_0x04_OServiceServiceRequest{}
	if err := oscar.Unmarshal(&snacPayloadIn, r); err != nil {
		return err
	}

	if snacPayloadIn.FoodGroup == CHAT {
		roomMeta, ok := snacPayloadIn.GetSlice(0x01)
		if !ok {
			return errors.New("missing room info")
		}

		roomSnac := oscar.SNAC_0x01_0x04_TLVRoomInfo{}
		if err := oscar.Unmarshal(&roomSnac, bytes.NewBuffer(roomMeta)); err != nil {
			return err
		}

		room, err := cr.Retrieve(string(roomSnac.Cookie))
		if err != nil {
			return sendInvalidSNACErr(snac, w, sequence)
		}
		room.NewSessionWithSN(sess.ID, sess.ScreenName)

		return sendChatRoomServiceInfo(cfg, room, sess, snac, sequence, w)
	}

	return sendInvalidSNACErr(snac, w, sequence)
}

func sendChatRoomServiceInfo(cfg Config, room ChatRoom, sess *Session, snac oscar.SnacFrame, sequence *uint32, w io.Writer) error {
	snacFrameOut := oscar.SnacFrame{
		FoodGroup: OSERVICE,
		SubGroup:  OServiceServiceResponse,
	}
	snacPayloadOut := oscar.SNAC_0x01_0x05_OServiceServiceResponse{
		TLVRestBlock: oscar.TLVRestBlock{
			TLVList: oscar.TLVList{
				{
					TType: OserviceTlvTagsReconnectHere,
					Val:   Address(cfg.OSCARHost, cfg.ChatPort),
				},
				{
					TType: OserviceTlvTagsLoginCookie,
					Val: ChatCookie{
						Cookie: []byte(room.Cookie),
						SessID: sess.ID,
					},
				},
				{
					TType: OserviceTlvTagsGroupId,
					Val:   CHAT,
				},
				{
					TType: OserviceTlvTagsSslCertname,
					Val:   "",
				},
				{
					TType: OserviceTlvTagsSslState,
					Val:   uint8(0x00),
				},
			},
		},
	}
	return writeOutSNAC(snac, snacFrameOut, snacPayloadOut, sequence, w)
}
