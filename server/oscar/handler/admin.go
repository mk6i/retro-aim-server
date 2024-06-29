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

func NewAdminHandler(logger *slog.Logger, adminService AdminService) AdminHandler {
	return AdminHandler{
		RouteLogger: middleware.RouteLogger{
			Logger: logger,
		},
		AdminService: adminService,
	}
}

type AdminHandler struct {
	AdminService
	OServiceService
	middleware.RouteLogger
}

type AdminService interface {
	ConfirmRequest(_ context.Context, frame wire.SNACFrame) (wire.SNACMessage, error)
	InfoQuery(ctx context.Context, sess *state.Session, frame wire.SNACFrame, body wire.SNAC_0x07_0x02_AdminInfoQuery) (wire.SNACMessage, error)
	InfoChangeRequest(ctx context.Context, sess *state.Session, frame wire.SNACFrame, body wire.SNAC_0x07_0x04_AdminInfoChangeRequest) (wire.SNACMessage, error)
}

func (rt AdminHandler) ConfirmRequest(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, _ io.Reader, rw oscar.ResponseWriter) error {
	outSNAC, err := rt.AdminService.ConfirmRequest(ctx, inFrame)
	if err != nil {
		return err
	}
	rt.LogRequestAndResponse(ctx, inFrame, nil, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (rt AdminHandler) InfoQuery(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x07_0x02_AdminInfoQuery{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	outSNAC, err := rt.AdminService.InfoQuery(ctx, sess, inFrame, inBody)
	if err != nil {
		return err
	}
	rt.LogRequestAndResponse(ctx, inFrame, nil, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (rt AdminHandler) InfoChangeRequest(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x07_0x04_AdminInfoChangeRequest{}
	if err := wire.Unmarshal(&inBody, r); err != nil {
		return err
	}
	outSNAC, err := rt.AdminService.InfoChangeRequest(ctx, sess, inFrame, inBody)
	if err != nil {
		return err
	}
	rt.LogRequestAndResponse(ctx, inFrame, nil, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}
