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

type ICQService interface {
	DBQuery(ctx context.Context, sess *state.Session, frame wire.SNACFrame, body wire.SNAC_0x0F_0x02_ICQDBQuery) error
}

func NewICQHandler(logger *slog.Logger, ICQService ICQService) ICQHandler {
	return ICQHandler{
		RouteLogger: middleware.RouteLogger{
			Logger: logger,
		},
		ICQService: ICQService,
	}
}

type ICQHandler struct {
	ICQService
	middleware.RouteLogger
}

func (rt ICQHandler) DBQuery(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x0F_0x02_ICQDBQuery{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}

	err := rt.ICQService.DBQuery(ctx, sess, inFrame, inBody)
	if err != nil {
		return err
	}
	//rt.LogRequest(ctx, inFrame, inBody)
	return nil
}
