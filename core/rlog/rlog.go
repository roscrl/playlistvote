package rlog

import (
	"context"
	"os"

	"app/core/contextkey"
	"golang.org/x/exp/slog"
)

func L(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(contextkey.RequestLogger{}).(*slog.Logger); ok {
		return logger
	}

	return NewDefaultLogger()
}

func NewDefaultLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}
