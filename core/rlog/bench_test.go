package rlog

import (
	"context"
	"os"
	"testing"

	"app/core/contextkey"
	"golang.org/x/exp/slog"
)

var (
	resultLog *slog.Logger
	resultCtx context.Context
)

func BenchmarkNewLoggerEveryRequest(b *testing.B) {
	b.ReportAllocs()

	var (
		logger *slog.Logger
		ctx    context.Context
	)

	for i := 0; i < b.N; i++ {
		ctx = context.Background()
		requestID := "123"

		textHandler := slog.NewTextHandler(os.Stdout, nil)
		requestContextHandler := ContextRequestHandler{Handler: textHandler}

		logger = slog.New(requestContextHandler)

		ctx = context.WithValue(ctx, contextkey.RequestLogger{}, logger)
		ctx = context.WithValue(ctx, contextkey.RequestID{}, requestID)
	}

	resultCtx = ctx
	resultLog = logger
}

func BenchmarkCloningWithLoggerEveryRequest(b *testing.B) {
	b.ReportAllocs()

	var (
		logger *slog.Logger
		ctx    context.Context
	)

	textHandler := slog.NewTextHandler(os.Stdout, nil)
	requestContextHandler := ContextRequestHandler{Handler: textHandler}
	logger = slog.New(requestContextHandler)

	for i := 0; i < b.N; i++ {
		ctx = context.Background()
		requestID := "123"

		logger2 := logger.With(slog.String(RequestIDLogKey, requestID))

		ctx = context.WithValue(ctx, contextkey.RequestLogger{}, logger2)
	}

	resultCtx = ctx
	resultLog = logger
}
