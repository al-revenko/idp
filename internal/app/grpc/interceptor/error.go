package interceptor

import (
	"context"
	"errors"
	"log/slog"

	"github.com/al-revenko/idp/internal/domain"
	"github.com/al-revenko/idp/internal/lib/valid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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

		if e, ok := errors.AsType[domain.DomainError](err); ok && e.Code == domain.CodeInternal {
			if e.Cause != nil {
				l.Error("internal error", slog.String("error", e.Error()), slog.String("cause", e.Cause.Error()))
			} else {
				l.Error("internal error", slog.String("error", e.Error()))
			}
		}

		grpcErr := errToGRPCStatus(err)

		return resp, grpcErr
	}
}

func errToGRPCStatus(err error) error {
	if domainErr, ok := errors.AsType[domain.DomainError](err); ok {
		switch domainErr.Code {
		case domain.CodeInvalidInput:
			return status.Error(codes.InvalidArgument, domainErr.Msg)
		case domain.CodeNotFound:
			return status.Error(codes.NotFound, domainErr.Msg)
		case domain.CodeConflict:
			return status.Error(codes.AlreadyExists, domainErr.Msg)
		case domain.CodeForbidden:
			return status.Error(codes.PermissionDenied, domainErr.Msg)
		case domain.CodeUnauthorized:
			return status.Error(codes.Unauthenticated, domainErr.Msg)
		}
	}

	if validationErr, ok := errors.AsType[*valid.ValidationError](err); ok {
		return status.Error(codes.InvalidArgument, validationErr.Error())
	}

	return status.Error(codes.Internal, "internal error")
}
