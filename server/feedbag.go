package server

import (
	"errors"
	"io"
	"time"

	"github.com/mkaminski/goaim/oscar"
)

type FeedbagHandler interface {
	DeleteItemHandler(sm SessionManager, sess *Session, fm FeedbagManager, snacPayloadIn oscar.SNAC_0x13_0x0A_FeedbagDeleteItem) (XMessage, error)
	InsertItemHandler(sm SessionManager, sess *Session, fm FeedbagManager, snacPayloadIn oscar.SNAC_0x13_0x08_FeedbagInsertItem) (XMessage, error)
	QueryHandler(sess *Session, fm FeedbagManager) (XMessage, error)
	QueryIfModifiedHandler(sess *Session, fm FeedbagManager, snacPayloadIn oscar.SNAC_0x13_0x05_FeedbagQueryIfModified) (XMessage, error)
	RightsQueryHandler() XMessage
	StartClusterHandler(oscar.SNAC_0x13_0x11_FeedbagStartCluster)
	UpdateItemHandler(sm SessionManager, sess *Session, fm FeedbagManager, snacPayloadIn oscar.SNAC_0x13_0x09_FeedbagUpdateItem) (XMessage, error)
}

func NewFeedbagRouter() FeedbagRouter {
	return FeedbagRouter{
		FeedbagHandler: FeedbagService{},
	}
}

type FeedbagRouter struct {
	FeedbagHandler
}

func (rt FeedbagRouter) RouteFeedbag(sm SessionManager, sess *Session, fm FeedbagManager, snac oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch snac.SubGroup {
	case oscar.FeedbagRightsQuery:
		inSNAC := oscar.SNAC_0x13_0x02_FeedbagRightsQuery{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC := rt.RightsQueryHandler()
		return writeOutSNAC(snac, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	case oscar.FeedbagQuery:
		inSNAC, err := rt.QueryHandler(sess, fm)
		if err != nil {
			return err
		}
		return writeOutSNAC(snac, inSNAC.snacFrame, inSNAC.snacOut, sequence, w)
	case oscar.FeedbagQueryIfModified:
		inSNAC := oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.QueryIfModifiedHandler(sess, fm, inSNAC)
		if err != nil {
			return err
		}
		return writeOutSNAC(snac, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	case oscar.FeedbagUse:
		return nil
	case oscar.FeedbagInsertItem:
		inSNAC := oscar.SNAC_0x13_0x08_FeedbagInsertItem{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.InsertItemHandler(sm, sess, fm, inSNAC)
		if err != nil {
			return err
		}
		return writeOutSNAC(snac, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	case oscar.FeedbagUpdateItem:
		inSNAC := oscar.SNAC_0x13_0x09_FeedbagUpdateItem{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.UpdateItemHandler(sm, sess, fm, inSNAC)
		if err != nil {
			return err
		}
		return writeOutSNAC(snac, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	case oscar.FeedbagDeleteItem:
		inSNAC := oscar.SNAC_0x13_0x0A_FeedbagDeleteItem{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.DeleteItemHandler(sm, sess, fm, inSNAC)
		if err != nil {
			return err
		}
		return writeOutSNAC(snac, outSNAC.snacFrame, outSNAC.snacOut, sequence, w)
	case oscar.FeedbagStartCluster:
		inSNAC := oscar.SNAC_0x13_0x11_FeedbagStartCluster{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		rt.StartClusterHandler(inSNAC)
		return nil
	case oscar.FeedbagEndCluster:
		return nil
	default:
		return ErrUnsupportedSubGroup
	}
}

type FeedbagService struct {
}

func (s FeedbagService) RightsQueryHandler() XMessage {
	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: FEEDBAG,
			SubGroup:  oscar.FeedbagRightsReply,
		},
		snacOut: oscar.SNAC_0x13_0x03_FeedbagRightsReply{
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
		},
	}
}

func (s FeedbagService) QueryHandler(sess *Session, fm FeedbagManager) (XMessage, error) {
	fb, err := fm.Retrieve(sess.ScreenName)
	if err != nil {
		return XMessage{}, err
	}

	lm := time.UnixMilli(0)

	if len(fb) > 0 {
		lm, err = fm.LastModified(sess.ScreenName)
		if err != nil {
			return XMessage{}, err
		}
	}

	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: FEEDBAG,
			SubGroup:  oscar.FeedbagReply,
		},
		snacOut: oscar.SNAC_0x13_0x06_FeedbagReply{
			Version:    0,
			Items:      fb,
			LastUpdate: uint32(lm.Unix()),
		},
	}, nil
}

func (s FeedbagService) QueryIfModifiedHandler(sess *Session, fm FeedbagManager, snacPayloadIn oscar.SNAC_0x13_0x05_FeedbagQueryIfModified) (XMessage, error) {
	fb, err := fm.Retrieve(sess.ScreenName)
	if err != nil {
		return XMessage{}, err
	}

	lm := time.UnixMilli(0)

	if len(fb) > 0 {
		lm, err = fm.LastModified(sess.ScreenName)
		if err != nil {
			return XMessage{}, err
		}
		if lm.Before(time.Unix(int64(snacPayloadIn.LastUpdate), 0)) {
			return XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: FEEDBAG,
					SubGroup:  oscar.FeedbagReplyNotModified,
				},
				snacOut: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
					LastUpdate: uint32(lm.Unix()),
					Count:      uint8(len(fb)),
				},
			}, nil
		}
	}

	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: FEEDBAG,
			SubGroup:  oscar.FeedbagReply,
		},
		snacOut: oscar.SNAC_0x13_0x06_FeedbagReply{
			Version:    0,
			Items:      fb,
			LastUpdate: uint32(lm.Unix()),
		},
	}, nil
}

func (s FeedbagService) InsertItemHandler(sm SessionManager, sess *Session, fm FeedbagManager, snacPayloadIn oscar.SNAC_0x13_0x08_FeedbagInsertItem) (XMessage, error) {
	for _, item := range snacPayloadIn.Items {
		// don't let users block themselves, it causes the AIM client to go
		// into a weird state.
		if item.ClassID == 3 && item.Name == sess.ScreenName {
			return XMessage{
				snacFrame: oscar.SnacFrame{
					FoodGroup: FEEDBAG,
					SubGroup:  oscar.FeedbagErr,
				},
				snacOut: oscar.SnacError{
					Code: ErrorCodeNotSupportedByHost,
				},
			}, nil
		}
	}

	if err := fm.Upsert(sess.ScreenName, snacPayloadIn.Items); err != nil {
		return XMessage{}, nil
	}

	for _, item := range snacPayloadIn.Items {
		switch item.ClassID {
		case oscar.FeedbagClassIdBuddy, oscar.FeedbagClassIDPermit: // add new buddy
			err := NotifyBuddyArrived(item.Name, sess.ScreenName, sm)
			switch {
			case errors.Is(err, ErrSessNotFound):
				continue
			case err != nil:
				return XMessage{}, err
			}
		case oscar.FeedbagClassIDDeny: // block buddy
			// notify this user that buddy is offline
			err := NotifyBuddyDeparted(item.Name, sess.ScreenName, sm)
			switch {
			case errors.Is(err, ErrSessNotFound):
				continue
			case err != nil:
				return XMessage{}, err
			}
			// notify former buddy that this user is offline
			if err := NotifyBuddyDeparted(sess.ScreenName, item.Name, sm); err != nil {
				return XMessage{}, err
			}
		}
	}

	snacPayloadOut := oscar.SNAC_0x13_0x0E_FeedbagStatus{}
	for range snacPayloadIn.Items {
		snacPayloadOut.Results = append(snacPayloadOut.Results, 0x0000)
	}

	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: FEEDBAG,
			SubGroup:  oscar.FeedbagStatus,
		},
		snacOut: snacPayloadOut,
	}, nil
}

func (s FeedbagService) UpdateItemHandler(sm SessionManager, sess *Session, fm FeedbagManager, snacPayloadIn oscar.SNAC_0x13_0x09_FeedbagUpdateItem) (XMessage, error) {
	if err := fm.Upsert(sess.ScreenName, snacPayloadIn.Items); err != nil {
		return XMessage{}, nil
	}

	for _, item := range snacPayloadIn.Items {
		switch item.ClassID {
		case oscar.FeedbagClassIdBuddy, oscar.FeedbagClassIDPermit:
			err := NotifyBuddyArrived(item.Name, sess.ScreenName, sm)
			switch {
			case errors.Is(err, ErrSessNotFound):
				continue
			case err != nil:
				return XMessage{}, err
			}
		}
	}

	snacPayloadOut := oscar.SNAC_0x13_0x0E_FeedbagStatus{}
	for range snacPayloadIn.Items {
		snacPayloadOut.Results = append(snacPayloadOut.Results, 0x0000)
	}

	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: FEEDBAG,
			SubGroup:  oscar.FeedbagStatus,
		},
		snacOut: snacPayloadOut,
	}, nil
}

func (s FeedbagService) DeleteItemHandler(sm SessionManager, sess *Session, fm FeedbagManager, snacPayloadIn oscar.SNAC_0x13_0x0A_FeedbagDeleteItem) (XMessage, error) {
	if err := fm.Delete(sess.ScreenName, snacPayloadIn.Items); err != nil {
		return XMessage{}, err
	}

	for _, item := range snacPayloadIn.Items {
		if item.ClassID == oscar.FeedbagClassIDDeny {
			err := NotifyBuddyArrived(item.Name, sess.ScreenName, sm)
			switch {
			case errors.Is(err, ErrSessNotFound):
				continue
			case err != nil:
				return XMessage{}, err
			}
			err = NotifyBuddyArrived(sess.ScreenName, item.Name, sm)
			switch {
			case errors.Is(err, ErrSessNotFound):
				continue
			case err != nil:
				return XMessage{}, err
			}
		}
	}

	snacPayloadOut := oscar.SNAC_0x13_0x0E_FeedbagStatus{}
	for range snacPayloadIn.Items {
		snacPayloadOut.Results = append(snacPayloadOut.Results, 0x0000) // success by default
	}

	return XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: FEEDBAG,
			SubGroup:  oscar.FeedbagStatus,
		},
		snacOut: snacPayloadOut,
	}, nil
}

// StartClusterHandler exists to capture the SNAC input in unit tests to verify
// it's correctly unmarshalled.
func (s FeedbagService) StartClusterHandler(oscar.SNAC_0x13_0x11_FeedbagStartCluster) {
}

func NotifyBuddyArrived(screenNameFrom, screenNameTo string, sm SessionManager) error {
	sess, err := sm.RetrieveByScreenName(screenNameFrom)
	switch {
	case err != nil:
		return err
	case sess.Invisible(): // don't tell user this buddy is online
		return nil
	}
	sm.SendToScreenName(screenNameTo, XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: BUDDY,
			SubGroup:  BuddyArrived,
		},
		snacOut: oscar.SNAC_0x03_0x0A_BuddyArrived{
			TLVUserInfo: sess.GetTLVUserInfo(),
		},
	})

	return nil
}

func NotifyBuddyDeparted(screenNameFrom, screenNameTo string, sm SessionManager) error {
	sess, err := sm.RetrieveByScreenName(screenNameFrom)
	switch {
	case err != nil:
		return err
	case sess.Invisible(): // don't tell user this buddy is online
		return nil
	}

	sm.SendToScreenName(screenNameTo, XMessage{
		snacFrame: oscar.SnacFrame{
			FoodGroup: BUDDY,
			SubGroup:  BuddyDeparted,
		},
		snacOut: oscar.SNAC_0x03_0x0B_BuddyDeparted{
			TLVUserInfo: oscar.TLVUserInfo{
				// don't include the TLV block, otherwise the AIM client fails
				// to process the block event
				ScreenName:   sess.ScreenName,
				WarningLevel: sess.GetWarning(),
			},
		},
	})

	return nil
}
