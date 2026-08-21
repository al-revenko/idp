package interceptor

import (
	"context"
	"log/slog"
	"time"
	"uuid"

	"github.com/al-revenko/idp/internal/lib/derr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TraceInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		reqId := uuid.New().String()
		ctx = context.WithValue(ctx, "reqId", reqId)

		l := log.With(slog.String("reqId", reqId), slog.String("method", info.FullMethod))

		l.Info("REQ")

		start := time.Now()

		resp, err := handler(ctx, req)

		l = l.With(slog.Int64("ms", time.Since(start).Milliseconds()))

		status, ok := status.FromError(err)

		if err == nil {
			l.Info("RES", slog.String("status", "OK"), slog.Uint64("code", uint64(codes.OK)))
		}

		if err != nil {
			if ok {
				l.Info("RES", slog.String("status", status.Code().String()), slog.Uint64("code", uint64(status.Code())))
			} else {
				l.Warn("RES", slog.String("status", "unknown"))
			}
		}

		return resp, err
	}
}

func ErrorInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		reqId, ok := ctx.Value("reqId").(string)
		if !ok {
			reqId = "unknown"
		}

		l := log.With(slog.String("reqId", reqId))

		resp, err := handler(ctx, req)

		if err == nil {
			return resp, err
		}

		grpcErr := derr.ToGRPCStatus(err)

		if status, _ := status.FromError(grpcErr); status.Code() == codes.Internal {
			l.Error("internal error", slog.String("error", err.Error()))
		}

		return resp, grpcErr
	}
}
