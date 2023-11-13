package server

import (
	"context"
	"errors"
	"github.com/mkaminski/goaim/user"
	"io"
	"log/slog"
	"time"

	"github.com/mkaminski/goaim/oscar"
)

type FeedbagHandler interface {
	DeleteItemHandler(ctx context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x13_0x0A_FeedbagDeleteItem) (oscar.XMessage, error)
	InsertItemHandler(ctx context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x13_0x08_FeedbagInsertItem) (oscar.XMessage, error)
	QueryHandler(ctx context.Context, sess *user.Session) (oscar.XMessage, error)
	QueryIfModifiedHandler(ctx context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x13_0x05_FeedbagQueryIfModified) (oscar.XMessage, error)
	RightsQueryHandler(context.Context) oscar.XMessage
	StartClusterHandler(context.Context, oscar.SNAC_0x13_0x11_FeedbagStartCluster)
	UpdateItemHandler(ctx context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x13_0x09_FeedbagUpdateItem) (oscar.XMessage, error)
}

func NewFeedbagRouter(logger *slog.Logger, sm SessionManager, fm FeedbagManager) FeedbagRouter {
	return FeedbagRouter{
		FeedbagHandler: FeedbagService{
			sm: sm,
			fm: fm,
		},
		RouteLogger: RouteLogger{
			Logger: logger,
		},
	}
}

type FeedbagRouter struct {
	FeedbagHandler
	RouteLogger
}

func (rt FeedbagRouter) RouteFeedbag(ctx context.Context, sess *user.Session, SNACFrame oscar.SnacFrame, r io.Reader, w io.Writer, sequence *uint32) error {
	switch SNACFrame.SubGroup {
	case oscar.FeedbagRightsQuery:
		inSNAC := oscar.SNAC_0x13_0x02_FeedbagRightsQuery{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC := rt.RightsQueryHandler(ctx)
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.FeedbagQuery:
		inSNAC, err := rt.QueryHandler(ctx, sess)
		if err != nil {
			return err
		}
		rt.logRequest(ctx, SNACFrame, inSNAC)
		return writeOutSNAC(SNACFrame, inSNAC.SnacFrame, inSNAC.SnacOut, sequence, w)
	case oscar.FeedbagQueryIfModified:
		inSNAC := oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.QueryIfModifiedHandler(ctx, sess, inSNAC)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.FeedbagUse:
		rt.logRequest(ctx, SNACFrame, nil)
		return nil
	case oscar.FeedbagInsertItem:
		inSNAC := oscar.SNAC_0x13_0x08_FeedbagInsertItem{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.InsertItemHandler(ctx, sess, inSNAC)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.FeedbagUpdateItem:
		inSNAC := oscar.SNAC_0x13_0x09_FeedbagUpdateItem{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.UpdateItemHandler(ctx, sess, inSNAC)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.FeedbagDeleteItem:
		inSNAC := oscar.SNAC_0x13_0x0A_FeedbagDeleteItem{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		outSNAC, err := rt.DeleteItemHandler(ctx, sess, inSNAC)
		if err != nil {
			return err
		}
		rt.logRequestAndResponse(ctx, SNACFrame, inSNAC, outSNAC.SnacFrame, outSNAC.SnacOut)
		return writeOutSNAC(SNACFrame, outSNAC.SnacFrame, outSNAC.SnacOut, sequence, w)
	case oscar.FeedbagStartCluster:
		inSNAC := oscar.SNAC_0x13_0x11_FeedbagStartCluster{}
		if err := oscar.Unmarshal(&inSNAC, r); err != nil {
			return err
		}
		rt.StartClusterHandler(ctx, inSNAC)
		rt.logRequest(ctx, SNACFrame, inSNAC)
		return nil
	case oscar.FeedbagEndCluster:
		rt.logRequest(ctx, SNACFrame, nil)
		return nil
	default:
		return ErrUnsupportedSubGroup
	}
}

type FeedbagService struct {
	sm SessionManager
	fm FeedbagManager
}

func (s FeedbagService) RightsQueryHandler(context.Context) oscar.XMessage {
	return oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.FEEDBAG,
			SubGroup:  oscar.FeedbagRightsReply,
		},
		SnacOut: oscar.SNAC_0x13_0x03_FeedbagRightsReply{
			TLVRestBlock: oscar.TLVRestBlock{
				TLVList: oscar.TLVList{
					oscar.NewTLV(0x03, uint16(200)),
					oscar.NewTLV(0x04, []uint16{
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
					}),
					oscar.NewTLV(0x05, uint16(200)),
					oscar.NewTLV(0x06, uint16(200)),
					oscar.NewTLV(0x07, uint16(200)),
					oscar.NewTLV(0x08, uint16(200)),
					oscar.NewTLV(0x09, uint16(200)),
					oscar.NewTLV(0x0A, uint16(200)),
					oscar.NewTLV(0x0C, uint16(200)),
					oscar.NewTLV(0x0D, uint16(200)),
					oscar.NewTLV(0x0E, uint16(100)),
				},
			},
		},
	}
}

func (s FeedbagService) QueryHandler(_ context.Context, sess *user.Session) (oscar.XMessage, error) {
	fb, err := s.fm.Retrieve(sess.ScreenName())
	if err != nil {
		return oscar.XMessage{}, err
	}

	lm := time.UnixMilli(0)

	if len(fb) > 0 {
		lm, err = s.fm.LastModified(sess.ScreenName())
		if err != nil {
			return oscar.XMessage{}, err
		}
	}

	return oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.FEEDBAG,
			SubGroup:  oscar.FeedbagReply,
		},
		SnacOut: oscar.SNAC_0x13_0x06_FeedbagReply{
			Version:    0,
			Items:      fb,
			LastUpdate: uint32(lm.Unix()),
		},
	}, nil
}

func (s FeedbagService) QueryIfModifiedHandler(_ context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x13_0x05_FeedbagQueryIfModified) (oscar.XMessage, error) {
	fb, err := s.fm.Retrieve(sess.ScreenName())
	if err != nil {
		return oscar.XMessage{}, err
	}

	lm := time.UnixMilli(0)

	if len(fb) > 0 {
		lm, err = s.fm.LastModified(sess.ScreenName())
		if err != nil {
			return oscar.XMessage{}, err
		}
		if lm.Before(time.Unix(int64(snacPayloadIn.LastUpdate), 0)) {
			return oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagReplyNotModified,
				},
				SnacOut: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
					LastUpdate: uint32(lm.Unix()),
					Count:      uint8(len(fb)),
				},
			}, nil
		}
	}

	return oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.FEEDBAG,
			SubGroup:  oscar.FeedbagReply,
		},
		SnacOut: oscar.SNAC_0x13_0x06_FeedbagReply{
			Version:    0,
			Items:      fb,
			LastUpdate: uint32(lm.Unix()),
		},
	}, nil
}

func (s FeedbagService) InsertItemHandler(ctx context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x13_0x08_FeedbagInsertItem) (oscar.XMessage, error) {
	for _, item := range snacPayloadIn.Items {
		// don't let users block themselves, it causes the AIM client to go
		// into a weird state.
		if item.ClassID == 3 && item.Name == sess.ScreenName() {
			return oscar.XMessage{
				SnacFrame: oscar.SnacFrame{
					FoodGroup: oscar.FEEDBAG,
					SubGroup:  oscar.FeedbagErr,
				},
				SnacOut: oscar.SnacError{
					Code: oscar.ErrorCodeNotSupportedByHost,
				},
			}, nil
		}
	}

	if err := s.fm.Upsert(sess.ScreenName(), snacPayloadIn.Items); err != nil {
		return oscar.XMessage{}, nil
	}

	for _, item := range snacPayloadIn.Items {
		switch item.ClassID {
		case oscar.FeedbagClassIdBuddy, oscar.FeedbagClassIDPermit: // add new buddy
			err := UnicastArrival(ctx, item.Name, sess.ScreenName(), s.sm)
			switch {
			case errors.Is(err, ErrSessNotFound):
				continue
			case err != nil:
				return oscar.XMessage{}, err
			}
		case oscar.FeedbagClassIDDeny: // block buddy
			// notify this user that buddy is offline
			err := UnicastDeparture(ctx, item.Name, sess.ScreenName(), s.sm)
			switch {
			case errors.Is(err, ErrSessNotFound):
				continue
			case err != nil:
				return oscar.XMessage{}, err
			}
			// notify former buddy that this user is offline
			if err := UnicastDeparture(ctx, sess.ScreenName(), item.Name, s.sm); err != nil {
				return oscar.XMessage{}, err
			}
		}
	}

	snacPayloadOut := oscar.SNAC_0x13_0x0E_FeedbagStatus{}
	for range snacPayloadIn.Items {
		snacPayloadOut.Results = append(snacPayloadOut.Results, 0x0000)
	}

	return oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.FEEDBAG,
			SubGroup:  oscar.FeedbagStatus,
		},
		SnacOut: snacPayloadOut,
	}, nil
}

func (s FeedbagService) UpdateItemHandler(ctx context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x13_0x09_FeedbagUpdateItem) (oscar.XMessage, error) {
	if err := s.fm.Upsert(sess.ScreenName(), snacPayloadIn.Items); err != nil {
		return oscar.XMessage{}, nil
	}

	for _, item := range snacPayloadIn.Items {
		switch item.ClassID {
		case oscar.FeedbagClassIdBuddy, oscar.FeedbagClassIDPermit:
			err := UnicastArrival(ctx, item.Name, sess.ScreenName(), s.sm)
			switch {
			case errors.Is(err, ErrSessNotFound):
				continue
			case err != nil:
				return oscar.XMessage{}, err
			}
		}
	}

	snacPayloadOut := oscar.SNAC_0x13_0x0E_FeedbagStatus{}
	for range snacPayloadIn.Items {
		snacPayloadOut.Results = append(snacPayloadOut.Results, 0x0000)
	}

	return oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.FEEDBAG,
			SubGroup:  oscar.FeedbagStatus,
		},
		SnacOut: snacPayloadOut,
	}, nil
}

func (s FeedbagService) DeleteItemHandler(ctx context.Context, sess *user.Session, snacPayloadIn oscar.SNAC_0x13_0x0A_FeedbagDeleteItem) (oscar.XMessage, error) {
	if err := s.fm.Delete(sess.ScreenName(), snacPayloadIn.Items); err != nil {
		return oscar.XMessage{}, err
	}

	for _, item := range snacPayloadIn.Items {
		if item.ClassID == oscar.FeedbagClassIDDeny {
			err := UnicastArrival(ctx, item.Name, sess.ScreenName(), s.sm)
			switch {
			case errors.Is(err, ErrSessNotFound):
				continue
			case err != nil:
				return oscar.XMessage{}, err
			}
			err = UnicastArrival(ctx, sess.ScreenName(), item.Name, s.sm)
			switch {
			case errors.Is(err, ErrSessNotFound):
				continue
			case err != nil:
				return oscar.XMessage{}, err
			}
		}
	}

	snacPayloadOut := oscar.SNAC_0x13_0x0E_FeedbagStatus{}
	for range snacPayloadIn.Items {
		snacPayloadOut.Results = append(snacPayloadOut.Results, 0x0000) // success by default
	}

	return oscar.XMessage{
		SnacFrame: oscar.SnacFrame{
			FoodGroup: oscar.FEEDBAG,
			SubGroup:  oscar.FeedbagStatus,
		},
		SnacOut: snacPayloadOut,
	}, nil
}

// StartClusterHandler exists to capture the SNAC input in unit tests to verify
// it's correctly unmarshalled.
func (s FeedbagService) StartClusterHandler(context.Context, oscar.SNAC_0x13_0x11_FeedbagStartCluster) {
}
