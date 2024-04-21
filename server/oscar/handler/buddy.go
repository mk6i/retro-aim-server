package handler

import (
	"context"
	"io"
	"log/slog"

	"github.com/mk6i/retro-aim-server/server/oscar"
	"github.com/mk6i/retro-aim-server/server/oscar/middleware"
	"github.com/mk6i/retro-aim-server/state"
	"github.com/mk6i/retro-aim-server/wire"
)

type BuddyService interface {
	RightsQuery(ctx context.Context, inFrame wire.SNACFrame) wire.SNACMessage
	AddBuddies(ctx context.Context, sess *state.Session, inBody wire.SNAC_0x03_0x04_BuddyAddBuddies) error
	DelBuddies(_ context.Context, sess *state.Session, inBody wire.SNAC_0x03_0x05_BuddyDelBuddies)
}

func NewBuddyHandler(logger *slog.Logger, buddyService BuddyService) BuddyHandler {
	return BuddyHandler{
		BuddyService: buddyService,
		RouteLogger: middleware.RouteLogger{
			Logger: logger,
		},
	}
}

type BuddyHandler struct {
	BuddyService
	middleware.RouteLogger
}

func (rt BuddyHandler) RightsQuery(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inSNAC := wire.SNAC_0x03_0x02_BuddyRightsQuery{}
	if err := wire.Unmarshal(&inSNAC, r); err != nil {
		return err
	}
	outSNAC := rt.BuddyService.RightsQuery(ctx, inFrame)
	rt.LogRequestAndResponse(ctx, inFrame, inSNAC, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (rt BuddyHandler) AddBuddies(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inSNAC := wire.SNAC_0x03_0x04_BuddyAddBuddies{}
	if err := wire.Unmarshal(&inSNAC, r); err != nil {
		return err
	}
	rt.LogRequest(ctx, inFrame, inSNAC)
	return rt.BuddyService.AddBuddies(ctx, sess, inSNAC)
}

func (rt BuddyHandler) DelBuddies(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inSNAC := wire.SNAC_0x03_0x05_BuddyDelBuddies{}
	if err := wire.Unmarshal(&inSNAC, r); err != nil {
		return err
	}
	rt.LogRequest(ctx, inFrame, inSNAC)
	rt.BuddyService.DelBuddies(ctx, sess, inSNAC)
	return nil
}
