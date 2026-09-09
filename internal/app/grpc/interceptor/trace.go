package interceptor

import (
	"context"
	"log/slog"
	"time"
	"uuid"

	"github.com/al-revenko/idp/internal/lib/meta"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TraceInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		reqId := uuid.New().String()
		ctx = context.WithValue(ctx, meta.CtxReqIDKey{}, reqId)

		l := log.With(slog.String("reqId", reqId), slog.String("method", info.FullMethod))

		l.Info("REQ")

		resp, err := handler(ctx, req)

		l = l.With(slog.Int64("ms", time.Since(start).Milliseconds()))

		if err != nil {
			status, ok := status.FromError(err)
			if ok {
				l.Info("RES", slog.Uint64("code", uint64(status.Code())), slog.String("status", status.Code().String()))
			} else {
				l.Warn("RES", slog.String("status", "unknown"))
			}
		} else {
			l.Info("RES", slog.String("status", "OK"), slog.Uint64("code", uint64(codes.OK)))
		}

		return resp, err
	}
}
