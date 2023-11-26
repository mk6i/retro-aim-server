package handler

import (
	"context"
	"time"

	"github.com/mkaminski/goaim/oscar"
	"github.com/mkaminski/goaim/state"
)

func NewFeedbagService(messageRelayer MessageRelayer, feedbagManager FeedbagManager) *FeedbagService {
	return &FeedbagService{messageRelayer: messageRelayer, feedbagManager: feedbagManager}
}

type FeedbagService struct {
	messageRelayer MessageRelayer
	feedbagManager FeedbagManager
}

func (s FeedbagService) RightsQueryHandler(_ context.Context, inFrame oscar.SNACFrame) oscar.SNACMessage {
	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Feedbag,
			SubGroup:  oscar.FeedbagRightsReply,
			RequestID: inFrame.RequestID,
		},
		Body: oscar.SNAC_0x13_0x03_FeedbagRightsReply{
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

func (s FeedbagService) QueryHandler(_ context.Context, sess *state.Session, inFrame oscar.SNACFrame) (oscar.SNACMessage, error) {
	fb, err := s.feedbagManager.Retrieve(sess.ScreenName())
	if err != nil {
		return oscar.SNACMessage{}, err
	}

	lm := time.UnixMilli(0)

	if len(fb) > 0 {
		lm, err = s.feedbagManager.LastModified(sess.ScreenName())
		if err != nil {
			return oscar.SNACMessage{}, err
		}
	}

	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Feedbag,
			SubGroup:  oscar.FeedbagReply,
			RequestID: inFrame.RequestID,
		},
		Body: oscar.SNAC_0x13_0x06_FeedbagReply{
			Version:    0,
			Items:      fb,
			LastUpdate: uint32(lm.Unix()),
		},
	}, nil
}

func (s FeedbagService) QueryIfModifiedHandler(_ context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x13_0x05_FeedbagQueryIfModified) (oscar.SNACMessage, error) {
	fb, err := s.feedbagManager.Retrieve(sess.ScreenName())
	if err != nil {
		return oscar.SNACMessage{}, err
	}

	lm := time.UnixMilli(0)

	if len(fb) > 0 {
		lm, err = s.feedbagManager.LastModified(sess.ScreenName())
		if err != nil {
			return oscar.SNACMessage{}, err
		}
		if lm.Before(time.Unix(int64(inBody.LastUpdate), 0)) {
			return oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagReplyNotModified,
					RequestID: inFrame.RequestID,
				},
				Body: oscar.SNAC_0x13_0x05_FeedbagQueryIfModified{
					LastUpdate: uint32(lm.Unix()),
					Count:      uint8(len(fb)),
				},
			}, nil
		}
	}

	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Feedbag,
			SubGroup:  oscar.FeedbagReply,
			RequestID: inFrame.RequestID,
		},
		Body: oscar.SNAC_0x13_0x06_FeedbagReply{
			Version:    0,
			Items:      fb,
			LastUpdate: uint32(lm.Unix()),
		},
	}, nil
}

func (s FeedbagService) InsertItemHandler(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x13_0x08_FeedbagInsertItem) (oscar.SNACMessage, error) {
	for _, item := range inBody.Items {
		// don't let users block themselves, it causes the AIM client to go
		// into a weird state.
		if item.ClassID == 3 && item.Name == sess.ScreenName() {
			return oscar.SNACMessage{
				Frame: oscar.SNACFrame{
					FoodGroup: oscar.Feedbag,
					SubGroup:  oscar.FeedbagErr,
					RequestID: inFrame.RequestID,
				},
				Body: oscar.SNACError{
					Code: oscar.ErrorCodeNotSupportedByHost,
				},
			}, nil
		}
	}

	if err := s.feedbagManager.Upsert(sess.ScreenName(), inBody.Items); err != nil {
		return oscar.SNACMessage{}, nil
	}

	for _, item := range inBody.Items {
		switch item.ClassID {
		case oscar.FeedbagClassIdBuddy, oscar.FeedbagClassIDPermit: // add new buddy
			unicastArrival(ctx, item.Name, sess.ScreenName(), s.messageRelayer)
		case oscar.FeedbagClassIDDeny: // block buddy
			// notify this user that buddy is offline
			unicastDeparture(ctx, item.Name, sess.ScreenName(), s.messageRelayer)
			// notify former buddy that this user is offline
			unicastDeparture(ctx, sess.ScreenName(), item.Name, s.messageRelayer)
		}
	}

	snacPayloadOut := oscar.SNAC_0x13_0x0E_FeedbagStatus{}
	for range inBody.Items {
		snacPayloadOut.Results = append(snacPayloadOut.Results, 0x0000)
	}

	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Feedbag,
			SubGroup:  oscar.FeedbagStatus,
			RequestID: inFrame.RequestID,
		},
		Body: snacPayloadOut,
	}, nil
}

func (s FeedbagService) UpdateItemHandler(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x13_0x09_FeedbagUpdateItem) (oscar.SNACMessage, error) {
	if err := s.feedbagManager.Upsert(sess.ScreenName(), inBody.Items); err != nil {
		return oscar.SNACMessage{}, nil
	}

	for _, item := range inBody.Items {
		switch item.ClassID {
		case oscar.FeedbagClassIdBuddy, oscar.FeedbagClassIDPermit:
			unicastArrival(ctx, item.Name, sess.ScreenName(), s.messageRelayer)
		}
	}

	snacPayloadOut := oscar.SNAC_0x13_0x0E_FeedbagStatus{}
	for range inBody.Items {
		snacPayloadOut.Results = append(snacPayloadOut.Results, 0x0000)
	}

	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Feedbag,
			SubGroup:  oscar.FeedbagStatus,
			RequestID: inFrame.RequestID,
		},
		Body: snacPayloadOut,
	}, nil
}

func (s FeedbagService) DeleteItemHandler(ctx context.Context, sess *state.Session, inFrame oscar.SNACFrame, inBody oscar.SNAC_0x13_0x0A_FeedbagDeleteItem) (oscar.SNACMessage, error) {
	if err := s.feedbagManager.Delete(sess.ScreenName(), inBody.Items); err != nil {
		return oscar.SNACMessage{}, err
	}

	for _, item := range inBody.Items {
		if item.ClassID == oscar.FeedbagClassIDDeny {
			unicastArrival(ctx, item.Name, sess.ScreenName(), s.messageRelayer)
			unicastArrival(ctx, sess.ScreenName(), item.Name, s.messageRelayer)
		}
	}

	snacPayloadOut := oscar.SNAC_0x13_0x0E_FeedbagStatus{}
	for range inBody.Items {
		snacPayloadOut.Results = append(snacPayloadOut.Results, 0x0000) // success by default
	}

	return oscar.SNACMessage{
		Frame: oscar.SNACFrame{
			FoodGroup: oscar.Feedbag,
			SubGroup:  oscar.FeedbagStatus,
			RequestID: inFrame.RequestID,
		},
		Body: snacPayloadOut,
	}, nil
}

// StartClusterHandler exists to capture the SNAC input in unit tests to verify
// it's correctly unmarshalled.
func (s FeedbagService) StartClusterHandler(context.Context, oscar.SNACFrame, oscar.SNAC_0x13_0x11_FeedbagStartCluster) {
}
