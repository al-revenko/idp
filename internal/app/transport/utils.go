package transport

import (
	"context"
	"log/slog"
	"time"
	"uuid"

	"github.com/al-revenko/idp/internal/lib/meta"
)

type TraceParams struct {
	Ctx       context.Context
	Transport Transport
	Log       *slog.Logger
}

func Trace(params TraceParams, fn func(ctx context.Context, log *slog.Logger) *slog.Logger) {
	start := time.Now()

	reqId := uuid.New().String()
	ctxWithId := context.WithValue(params.Ctx, meta.CtxReqIDKey{}, reqId)

	l := params.Log.With(slog.String("reqId", reqId), slog.String("transport", string(params.Transport)))

	l.Info("REQ")

	l = fn(ctxWithId, l)

	l = l.With(slog.Int64("ms", time.Since(start).Milliseconds()))

	l.Info("RES")
}
