package server

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/mkaminski/goaim/oscar"
)

const (
	LevelTrace = slog.Level(-8)
)

var levelNames = map[slog.Leveler]string{
	LevelTrace: "TRACE",
}

func NewLogger(cfg Config) *slog.Logger {
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
	return slog.New(Handler{slog.NewTextHandler(os.Stdout, opts)})
}

type Handler struct {
	slog.Handler
}

func (h Handler) Handle(ctx context.Context, r slog.Record) error {
	if sn := ctx.Value("screenName"); sn != nil {
		r.AddAttrs(slog.Attr{Key: "screenName", Value: slog.StringValue(sn.(string))})
	}
	if ip := ctx.Value("ip"); ip != nil {
		r.AddAttrs(slog.Attr{Key: "ip", Value: slog.StringValue(ip.(string))})
	}
	return h.Handler.Handle(ctx, r)
}

func (h Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return Handler{h.Handler.WithAttrs(attrs)}
}

func (h Handler) WithGroup(name string) slog.Handler {
	return h.Handler.WithGroup(name)
}

type RouteLogger struct {
	Logger *slog.Logger
}

func (rt RouteLogger) logRequestAndResponse(ctx context.Context, inFrame oscar.SnacFrame, inSNAC any, outFrame oscar.SnacFrame, outSNAC any) {
	msg := "client request -> server response"
	switch {
	case rt.Logger.Enabled(ctx, LevelTrace):
		rt.Logger.LogAttrs(ctx, LevelTrace, msg, SNACLogGroupWithPayload("request", inFrame, inSNAC),
			SNACLogGroupWithPayload("response", outFrame, outSNAC))
	case rt.Logger.Enabled(ctx, slog.LevelDebug):
		rt.Logger.LogAttrs(ctx, slog.LevelDebug, msg, SNACLogGroup("request", inFrame),
			SNACLogGroup("response", outFrame))
	}
}

func (rt RouteLogger) logRequestError(ctx context.Context, inFrame oscar.SnacFrame, err error) {
	logRequestError(ctx, rt.Logger, inFrame, err)
}

func logRequestError(ctx context.Context, logger *slog.Logger, inFrame oscar.SnacFrame, err error) {
	logger.LogAttrs(ctx, slog.LevelError, "client request error",
		slog.Group("request",
			slog.String("food_group", oscar.FoodGroupStr(inFrame.FoodGroup)),
			slog.String("sub_group", oscar.SubGroupStr(inFrame.FoodGroup, inFrame.SubGroup)),
		),
		slog.String("err", err.Error()),
	)
}

func (rt RouteLogger) logRequest(ctx context.Context, inFrame oscar.SnacFrame, inSNAC any) {
	logRequest(ctx, rt.Logger, inFrame, inSNAC)
}

func logRequest(ctx context.Context, logger *slog.Logger, inFrame oscar.SnacFrame, inSNAC any) {
	const msg = "client request"
	switch {
	case logger.Enabled(ctx, LevelTrace):
		logger.LogAttrs(ctx, LevelTrace, msg, SNACLogGroupWithPayload("request", inFrame, inSNAC))
	case logger.Enabled(ctx, slog.LevelDebug):
		logger.LogAttrs(ctx, slog.LevelDebug, msg, slog.Group("request", SNACLogGroup("request", inFrame)))
	}
}

func SNACLogGroup(key string, outFrame oscar.SnacFrame) slog.Attr {
	return slog.Group(key,
		slog.String("food_group", oscar.FoodGroupStr(outFrame.FoodGroup)),
		slog.String("sub_group", oscar.SubGroupStr(outFrame.FoodGroup, outFrame.SubGroup)),
	)
}

func SNACLogGroupWithPayload(key string, outFrame oscar.SnacFrame, outSNAC any) slog.Attr {
	return slog.Group(key,
		slog.String("food_group", oscar.FoodGroupStr(outFrame.FoodGroup)),
		slog.String("sub_group", oscar.SubGroupStr(outFrame.FoodGroup, outFrame.SubGroup)),
		slog.Any("snac_frame", outFrame),
		slog.Any("snac_payload", outSNAC),
	)
}
