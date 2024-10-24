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

type UserLookupService interface {
	FindByEmail(ctx context.Context, inFrame wire.SNACFrame, inBody wire.SNAC_0x0A_0x02_UserLookupFindByEmail) (wire.SNACMessage, error)
}

func NewUserLookupHandler(logger *slog.Logger, userLookupService UserLookupService) UserLookupHandler {
	return UserLookupHandler{
		UserLookupService: userLookupService,
		RouteLogger: middleware.RouteLogger{
			Logger: logger,
		},
	}
}

type UserLookupHandler struct {
	UserLookupService
	middleware.RouteLogger
}

func (h UserLookupHandler) FindByEmail(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x0A_0x02_UserLookupFindByEmail{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}
	outSNAC, err := h.UserLookupService.FindByEmail(ctx, inFrame, inBody)
	if err != nil {
		return err
	}
	h.LogRequestAndResponse(ctx, inFrame, inBody, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}
