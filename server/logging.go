package server

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/mk6i/retro-aim-server/config"
	"github.com/mk6i/retro-aim-server/oscar"
)

const (
	LevelTrace = slog.Level(-8)
)

var levelNames = map[slog.Leveler]string{
	LevelTrace: "TRACE",
}

func NewLogger(cfg config.Config) *slog.Logger {
	var level slog.Level
	switch strings.ToLower(cfg.LogLevel) {
	case "trace":
		level = LevelTrace
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	case "info":
		fallthrough
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)
				levelLabel, exists := levelNames[level]
				if !exists {
					levelLabel = level.String()
				}
				a.Value = slog.StringValue(levelLabel)
			}

			return a
		},
	}
	return slog.New(handler{slog.NewTextHandler(os.Stdout, opts)})
}

type handler struct {
	slog.Handler
}

func (h handler) Handle(ctx context.Context, r slog.Record) error {
	if sn := ctx.Value("screenName"); sn != nil {
		r.AddAttrs(slog.Attr{Key: "screenName", Value: slog.StringValue(sn.(string))})
	}
	if ip := ctx.Value("ip"); ip != nil {
		r.AddAttrs(slog.Attr{Key: "ip", Value: slog.StringValue(ip.(string))})
	}
	return h.Handler.Handle(ctx, r)
}

func (h handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return handler{h.Handler.WithAttrs(attrs)}
}

func (h handler) WithGroup(name string) slog.Handler {
	return h.Handler.WithGroup(name)
}

type routeLogger struct {
	Logger *slog.Logger
}

func (rt routeLogger) logRequestAndResponse(ctx context.Context, inFrame oscar.SNACFrame, inSNAC any, outFrame oscar.SNACFrame, outSNAC any) {
	msg := "client request -> server response"
	switch {
	case rt.Logger.Enabled(ctx, LevelTrace):
		rt.Logger.LogAttrs(ctx, LevelTrace, msg, snacLogGroupWithPayload("request", inFrame, inSNAC),
			snacLogGroupWithPayload("response", outFrame, outSNAC))
	case rt.Logger.Enabled(ctx, slog.LevelDebug):
		rt.Logger.LogAttrs(ctx, slog.LevelDebug, msg, snacLogGroup("request", inFrame),
			snacLogGroup("response", outFrame))
	}
}

func (rt routeLogger) logRequestError(ctx context.Context, inFrame oscar.SNACFrame, err error) {
	logRequestError(ctx, rt.Logger, inFrame, err)
}

func logRequestError(ctx context.Context, logger *slog.Logger, inFrame oscar.SNACFrame, err error) {
	logger.LogAttrs(ctx, slog.LevelError, "client request error",
		slog.Group("request",
			slog.String("food_group", oscar.FoodGroupStr(inFrame.FoodGroup)),
			slog.String("sub_group", oscar.SubGroupStr(inFrame.FoodGroup, inFrame.SubGroup)),
		),
		slog.String("err", err.Error()),
	)
}

func (rt routeLogger) logRequest(ctx context.Context, inFrame oscar.SNACFrame, inSNAC any) {
	logRequest(ctx, rt.Logger, inFrame, inSNAC)
}

func logRequest(ctx context.Context, logger *slog.Logger, inFrame oscar.SNACFrame, inSNAC any) {
	const msg = "client request"
	switch {
	case logger.Enabled(ctx, LevelTrace):
		logger.LogAttrs(ctx, LevelTrace, msg, snacLogGroupWithPayload("request", inFrame, inSNAC))
	case logger.Enabled(ctx, slog.LevelDebug):
		logger.LogAttrs(ctx, slog.LevelDebug, msg, slog.Group("request", snacLogGroup("request", inFrame)))
	}
}

func snacLogGroup(key string, outFrame oscar.SNACFrame) slog.Attr {
	return slog.Group(key,
		slog.String("food_group", oscar.FoodGroupStr(outFrame.FoodGroup)),
		slog.String("sub_group", oscar.SubGroupStr(outFrame.FoodGroup, outFrame.SubGroup)),
	)
}

func snacLogGroupWithPayload(key string, outFrame oscar.SNACFrame, outSNAC any) slog.Attr {
	return slog.Group(key,
		slog.String("food_group", oscar.FoodGroupStr(outFrame.FoodGroup)),
		slog.String("sub_group", oscar.SubGroupStr(outFrame.FoodGroup, outFrame.SubGroup)),
		slog.Any("snac_frame", outFrame),
		slog.Any("snac_payload", outSNAC),
	)
}
