package server

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/mkaminski/goaim/oscar"
	"io"
	"time"
)

type OServiceHandler interface {
	ClientOnlineHandler(snacPayloadIn oscar.SNAC_0x01_0x02_OServiceClientOnline, onReadyCB OnReadyCB, sess *Session, sm SessionManager) ([]XMessage, error)
	ClientVersionsHandler(snacPayloadIn oscar.SNAC_0x01_0x17_OServiceClientVersions) XMessage
	IdleNotificationHandler(sess *Session, sm SessionManager, fm *FeedbagStore, snacPayloadIn oscar.SNAC_0x01_0x11_OServiceIdleNotification) error
	RateParamsQueryHandler() XMessage
	RateParamsSubAddHandler(oscar.SNAC_0x01_0x08_OServiceRateParamsSubAdd)
	ServiceRequestHandler(cfg Config, cr *ChatRegistry, sess *Session, snacPayloadIn oscar.SNAC_0x01_0x04_OServiceServiceRequest) (XMessage, error)
	SetUserInfoFieldsHandler(sess *Session, sm SessionManager, fm *FeedbagStore, snacPayloadIn oscar.SNAC_0x01_0x1E_OServiceSetUserInfoFields) (XMessage, error)
	UserInfoQueryHandler(sess *Session) XMessage
}

func NewOServiceRouter() OServiceRouter {
	return OServiceRouter{
		OServiceHandler: OServiceService{},
	}
}

type OServiceRouter struct {
	OServiceHandler
}

func (rt OServiceRouter) RouteOService(cfg Config, ready OnReadyCB, cr *ChatRegistry, sm SessionManager, fm *FeedbagStore, sess *Session, SNACFrame oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch SNACFrame.SubGroup {
	case oscar.OServiceClientOnline:
		inSNAC := oscar.SNAC_0x01_0x02_OServiceClientOnline{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		batch, err := rt.ClientOnlineHandler(inSNAC, ready, sess, sm)
		for _, msg := range batch {
			switch {
			case err != nil:
				return err
			case batch != nil:
				return writeOutSNAC(SNACFrame, msg.snacFrame, msg.snacOut, sequence, w)
			}
		}
		return nil
	case oscar.OServiceServiceRequest:
		inSNAC := oscar.SNAC_0x01_0x04_OServiceServiceRequest{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.ServiceRequestHandler(cfg, cr, sess, inSNAC)
		switch {
		case errors.Is(err, ErrUnsupportedSubGroup):
			return sendInvalidSNACErr(SNACFrame, w, sequence)
		case err != nil:
			return err
		}
		return writeOutSNAC(SNACFrame, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	case oscar.OServiceRateParamsQuery:
		outSNAC := rt.RateParamsQueryHandler()
		return writeOutSNAC(SNACFrame, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	case oscar.OServiceRateParamsSubAdd:
		inSNAC := oscar.SNAC_0x01_0x08_OServiceRateParamsSubAdd{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		rt.RateParamsSubAddHandler(inSNAC)
		return oscar.Unmarshal(&inSNAC, r)
	case oscar.OServiceUserInfoQuery:
		outSNAC := rt.UserInfoQueryHandler(sess)
		return writeOutSNAC(SNACFrame, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	case oscar.OServiceIdleNotification:
		inSNAC := oscar.SNAC_0x01_0x11_OServiceIdleNotification{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		return rt.IdleNotificationHandler(sess, sm, fm, inSNAC)
	case oscar.OServiceClientVersions:
		inSNAC := oscar.SNAC_0x01_0x17_OServiceClientVersions{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC := rt.ClientVersionsHandler(inSNAC)
		return writeOutSNAC(SNACFrame, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	case oscar.OServiceSetUserInfoFields:
		inSNAC := oscar.SNAC_0x01_0x1E_OServiceSetUserInfoFields{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.SetUserInfoFieldsHandler(sess, sm, fm, inSNAC)
		if err != nil {
			return err
		}
		return writeOutSNAC(SNACFrame, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	default:
		return ErrUnsupportedSubGroup
	}
}

type OServiceService struct {
}

func WriteOServiceHostOnline(foodGroups []uint16, w io.Writer, sequence *uint32) error {
	fmt.Println("writeOServiceHostOnline...")
	snacFrameOut := oscar.SnacFrame{
		FoodGroup: OSERVICE,
		SubGroup:  oscar.OServiceHostOnline,
	}
	snacPayloadOut := oscar.SNAC_0x01_0x03_OServiceHostOnline{
		FoodGroups: foodGroups,
	}
	return writeOutSNAC(oscar.SnacFrame{}, snacFrameOut, snacPayloadOut, sequence, w)
}

func (s OServiceService) ClientVersionsHandler(snacPayloadIn oscar.SNAC_0x01_0x17_OServiceClientVersions) XMessage {
	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: OSERVICE,
			SubGroup:  oscar.OServiceHostVersions,
		},
		snacOut: oscar.SNAC_0x01_0x18_OServiceHostVersions{
			Versions: snacPayloadIn.Versions,
		},
	}
}

func (s OServiceService) RateParamsQueryHandler() XMessage {
	snacFrameOut := oscar.SnacFrame{
		FoodGroup: OSERVICE,
		SubGroup:  oscar.OServiceRateParamsReply,
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

	return XMessage{
		snacFrame: snacFrameOut,
		snacOut:   snacPayloadOut,
	}
}

func (s OServiceService) UserInfoQueryHandler(sess *Session) XMessage {
	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: OSERVICE,
			SubGroup:  oscar.OServiceUserInfoUpdate,
		},
		snacOut: oscar.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
			TLVUserInfo: sess.GetTLVUserInfo(),
		},
	}
}

type OnReadyCB func(sess *Session, sm SessionManager) ([]XMessage, error)

func (s OServiceService) ClientOnlineHandler(snacPayloadIn oscar.SNAC_0x01_0x02_OServiceClientOnline, onReadyCB OnReadyCB, sess *Session, sm SessionManager) ([]XMessage, error) {
	for _, version := range snacPayloadIn.GroupVersions {
		fmt.Printf("ClientOnlineHandler read SNAC client messageType: %+v\n", version)
	}
	return onReadyCB(sess, sm)
}

func (s OServiceService) SetUserInfoFieldsHandler(sess *Session, sm SessionManager, fm *FeedbagStore, snacPayloadIn oscar.SNAC_0x01_0x1E_OServiceSetUserInfoFields) (XMessage, error) {
	if status, hasStatus := snacPayloadIn.GetUint32(0x06); hasStatus {
		switch status {
		case 0x000:
			sess.SetInvisible(false)
			if err := NotifyArrival(sess, sm, fm); err != nil {
				return XMessage{}, err
			}
		case 0x100:
			sess.SetInvisible(true)
			if err := NotifyDeparture(sess, sm, fm); err != nil {
				return XMessage{}, err
			}
		default:
			return XMessage{}, fmt.Errorf("don't know what to do with status %d", status)
		}
	}
	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: OSERVICE,
			SubGroup:  oscar.OServiceUserInfoUpdate,
		},
		snacOut: oscar.SNAC_0x01_0x0F_OServiceUserInfoUpdate{
			TLVUserInfo: sess.GetTLVUserInfo(),
		},
	}, nil
}

func (s OServiceService) IdleNotificationHandler(sess *Session, sm SessionManager, fm *FeedbagStore, snacPayloadIn oscar.SNAC_0x01_0x11_OServiceIdleNotification) error {
	if snacPayloadIn.IdleTime == 0 {
		sess.SetActive()
	} else {
		sess.SetIdle(time.Duration(snacPayloadIn.IdleTime) * time.Second)
	}
	return NotifyArrival(sess, sm, fm)
}

func (s OServiceService) ServiceRequestHandler(cfg Config, cr *ChatRegistry, sess *Session, snacPayloadIn oscar.SNAC_0x01_0x04_OServiceServiceRequest) (XMessage, error) {
	if snacPayloadIn.FoodGroup != CHAT {
		return XMessage{}, ErrUnsupportedSubGroup
	}

	roomMeta, ok := snacPayloadIn.GetSlice(0x01)
	if !ok {
		return XMessage{}, errors.New("missing room info")
	}

	roomSnac := oscar.SNAC_0x01_0x04_TLVRoomInfo{}
	if err := oscar.Unmarshal(&roomSnac, bytes.NewBuffer(roomMeta)); err != nil {
		return XMessage{}, err
	}

	room, err := cr.Retrieve(string(roomSnac.Cookie))
	if err != nil {
		return XMessage{}, ErrUnsupportedSubGroup
	}
	room.NewSessionWithSN(sess.ID, sess.ScreenName)

	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: OSERVICE,
			SubGroup:  oscar.OServiceServiceResponse,
		},
		snacOut: oscar.SNAC_0x01_0x05_OServiceServiceResponse{
			TLVRestBlock: oscar.TLVRestBlock{
				TLVList: oscar.TLVList{
					{
						TType: oscar.OServiceTLVTagsReconnectHere,
						Val:   Address(cfg.OSCARHost, cfg.ChatPort),
					},
					{
						TType: oscar.OServiceTLVTagsLoginCookie,
						Val: ChatCookie{
							Cookie: []byte(room.Cookie),
							SessID: sess.ID,
						},
					},
					{
						TType: oscar.OServiceTLVTagsGroupID,
						Val:   CHAT,
					},
					{
						TType: oscar.OServiceTLVTagsSSLCertName,
						Val:   "",
					},
					{
						TType: oscar.OServiceTLVTagsSSLState,
						Val:   uint8(0x00),
					},
				},
			},
		},
	}, nil
}

// RateParamsSubAddHandler exists to capture the SNAC input in unit tests to
// verify it's correctly unmarshalled.
func (s OServiceService) RateParamsSubAddHandler(oscar.SNAC_0x01_0x08_OServiceRateParamsSubAdd) {
}
