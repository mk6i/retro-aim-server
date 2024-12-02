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

type PermitDenyService interface {
	AddDenyListEntries(ctx context.Context, sess *state.Session, body wire.SNAC_0x09_0x07_PermitDenyAddDenyListEntries) error
	AddPermListEntries(ctx context.Context, sess *state.Session, body wire.SNAC_0x09_0x05_PermitDenyAddPermListEntries) error
	DelDenyListEntries(ctx context.Context, sess *state.Session, body wire.SNAC_0x09_0x08_PermitDenyDelDenyListEntries) error
	DelPermListEntries(ctx context.Context, sess *state.Session, body wire.SNAC_0x09_0x06_PermitDenyDelPermListEntries) error
	RightsQuery(_ context.Context, frame wire.SNACFrame) wire.SNACMessage
}

func NewPermitDenyHandler(logger *slog.Logger, permitDenyService PermitDenyService) PermitDenyHandler {
	return PermitDenyHandler{
		RouteLogger: middleware.RouteLogger{
			Logger: logger,
		},
		PermitDenyService: permitDenyService,
	}
}

type PermitDenyHandler struct {
	PermitDenyService
	middleware.RouteLogger
}

func (rt PermitDenyHandler) RightsQuery(ctx context.Context, _ *state.Session, inFrame wire.SNACFrame, _ io.Reader, rw oscar.ResponseWriter) error {
	outSNAC := rt.PermitDenyService.RightsQuery(ctx, inFrame)
	rt.LogRequestAndResponse(ctx, inFrame, nil, outSNAC.Frame, outSNAC.Body)
	return rw.SendSNAC(outSNAC.Frame, outSNAC.Body)
}

func (rt PermitDenyHandler) AddDenyListEntries(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x09_0x07_PermitDenyAddDenyListEntries{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}
	rt.LogRequest(ctx, inFrame, inBody)
	return rt.PermitDenyService.AddDenyListEntries(ctx, sess, inBody)
}

func (rt PermitDenyHandler) DelDenyListEntries(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x09_0x08_PermitDenyDelDenyListEntries{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}
	rt.LogRequest(ctx, inFrame, inBody)
	return rt.PermitDenyService.DelDenyListEntries(ctx, sess, inBody)
}

func (rt PermitDenyHandler) AddPermListEntries(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x09_0x05_PermitDenyAddPermListEntries{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}
	rt.LogRequest(ctx, inFrame, inBody)
	return rt.PermitDenyService.AddPermListEntries(ctx, sess, inBody)
}

func (rt PermitDenyHandler) DelPermListEntries(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x09_0x06_PermitDenyDelPermListEntries{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}
	rt.LogRequest(ctx, inFrame, inBody)
	return rt.PermitDenyService.DelPermListEntries(ctx, sess, inBody)
}

// SetGroupPermitMask sets the classes of users I can interact with. We don't
// apply any of these settings to the privacy mechanism, so just log them for
// now.
func (rt PermitDenyHandler) SetGroupPermitMask(ctx context.Context, sess *state.Session, inFrame wire.SNACFrame, r io.Reader, rw oscar.ResponseWriter) error {
	inBody := wire.SNAC_0x09_0x04_PermitDenySetGroupPermitMask{}
	if err := wire.UnmarshalBE(&inBody, r); err != nil {
		return err
	}

	var flags []string

	if inBody.IsFlagSet(wire.OServiceUserFlagUnconfirmed) {
		flags = append(flags, "wire.OServiceUserFlagUnconfirmed")
	}
	if inBody.IsFlagSet(wire.OServiceUserFlagAdministrator) {
		flags = append(flags, "wire.OServiceUserFlagAdministrator")
	}
	if inBody.IsFlagSet(wire.OServiceUserFlagAOL) {
		flags = append(flags, "wire.OServiceUserFlagAOL")
	}
	if inBody.IsFlagSet(wire.OServiceUserFlagOSCARPay) {
		flags = append(flags, "wire.OServiceUserFlagOSCARPay")
	}
	if inBody.IsFlagSet(wire.OServiceUserFlagOSCARFree) {
		flags = append(flags, "wire.OServiceUserFlagOSCARFree")
	}
	if inBody.IsFlagSet(wire.OServiceUserFlagUnavailable) {
		flags = append(flags, "wire.OServiceUserFlagUnavailable")
	}
	if inBody.IsFlagSet(wire.OServiceUserFlagICQ) {
		flags = append(flags, "wire.OServiceUserFlagICQ")
	}
	if inBody.IsFlagSet(wire.OServiceUserFlagWireless) {
		flags = append(flags, "wire.OServiceUserFlagWireless")
	}
	if inBody.IsFlagSet(wire.OServiceUserFlagInternal) {
		flags = append(flags, "wire.OServiceUserFlagInternal")
	}
	if inBody.IsFlagSet(wire.OServiceUserFlagFish) {
		flags = append(flags, "wire.OServiceUserFlagFish")
	}
	if inBody.IsFlagSet(wire.OServiceUserFlagBot) {
		flags = append(flags, "wire.OServiceUserFlagBot")
	}
	if inBody.IsFlagSet(wire.OServiceUserFlagBeast) {
		flags = append(flags, "wire.OServiceUserFlagBeast")
	}
	if inBody.IsFlagSet(wire.OServiceUserFlagOneWayWireless) {
		flags = append(flags, "wire.OServiceUserFlagOneWayWireless")
	}
	if inBody.IsFlagSet(wire.OServiceUserFlagOfficial) {
		flags = append(flags, "wire.OServiceUserFlagOfficial")
	}

	rt.Logger.Info("set pd group mask", "flags", flags)
	rt.LogRequest(ctx, inFrame, inBody)

	return nil
}
