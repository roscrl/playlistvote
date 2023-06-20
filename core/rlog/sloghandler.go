package rlog

import (
	"context"

	"app/core/contextkey"
	"golang.org/x/exp/slog"
)

type ContextRequestHandler struct {
	slog.Handler
}

const (
	RequestPathLogKey = "request_path"
	RequestIDLogKey   = "request_id"
	SessionIDLogKey   = "session_id"
)

func (h ContextRequestHandler) Handle(ctx context.Context, record slog.Record) error {
	if path, ok := ctx.Value(contextkey.RequestPath{}).(string); ok {
		record.AddAttrs(slog.String(RequestPathLogKey, path))
	}

	if rid, ok := ctx.Value(contextkey.RequestID{}).(string); ok {
		record.AddAttrs(slog.String(RequestIDLogKey, rid))
	}

	if sid, ok := ctx.Value(contextkey.SessionID{}).(string); ok {
		record.AddAttrs(slog.String(SessionIDLogKey, sid))
	}

	return h.Handler.Handle(ctx, record)
}
