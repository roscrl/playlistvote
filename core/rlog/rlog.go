package rlog

import (
	"context"
	"os"

	"golang.org/x/exp/slog"
)

type ContextRequestHandler struct {
	slog.Handler
}

type ContextKeyRequestLogger struct{}

type ContextKeyRequestID struct{}

const ContextKeyRequestIDLogKey = "request_id"

func (h ContextRequestHandler) Handle(ctx context.Context, r slog.Record) error {
	if rid, ok := ctx.Value(ContextKeyRequestID{}).(string); ok {
		r.AddAttrs(slog.String(ContextKeyRequestIDLogKey, rid))
	}

	return h.Handler.Handle(ctx, r)
}

func L(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(ContextKeyRequestLogger{}).(*slog.Logger); ok {
		return logger
	}

	return DefaultLogger()
}

func DefaultLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}
