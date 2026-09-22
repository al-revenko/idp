package interceptor

import (
	"context"
	"log/slog"

	"github.com/al-revenko/idp/internal/app/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TraceInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		traceLog := log.With(slog.String("method", info.FullMethod))

		var resp any
		var respErr error

		transport.Trace(transport.TraceParams{
			Ctx:       ctx,
			Log:       traceLog,
			Transport: transport.GRPC,
		}, func(ctx context.Context, log *slog.Logger) *slog.Logger {
			resp, respErr = handler(ctx, req)

			if respErr != nil {
				status, ok := status.FromError(respErr)
				if ok {
					return log.With(slog.Uint64("code", uint64(status.Code())), slog.String("status", status.Code().String()))
				} else {
					return log.With(slog.Uint64("code", uint64(codes.Unknown)), slog.String("status", "unknown"))
				}
			}

			return log.With(slog.String("status", "OK"), slog.Uint64("code", uint64(codes.OK)))
		})

		return resp, respErr
	}
}
