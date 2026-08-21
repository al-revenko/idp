package derr

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	CodeInvalidInput = iota
	CodeNotFound
	CodeConflict
	CodeForbidden
	CodeUnauthorized
	CodeInternal
)

type DomainError struct {
	Code  int32
	Msg   string
	Cause error
}

func (e DomainError) Error() string {
	return e.Msg
}

func (e DomainError) Unwrap() error {
	if e.Cause == nil {
		return nil
	}

	return e.Cause
}

func New(code int32, msg string) DomainError {
	return DomainError{Code: code, Msg: msg}
}

func From(code int32, msg string, err error) DomainError {
	return DomainError{Code: code, Msg: msg, Cause: err}
}

func ToGRPCStatus(err error) error {
	if domainErr, ok := errors.AsType[DomainError](err); ok {
		switch domainErr.Code {
		case CodeInvalidInput:
			return status.Error(codes.InvalidArgument, domainErr.Msg)
		case CodeNotFound:
			return status.Error(codes.NotFound, domainErr.Msg)
		case CodeConflict:
			return status.Error(codes.AlreadyExists, domainErr.Msg)
		case CodeForbidden:
			return status.Error(codes.PermissionDenied, domainErr.Msg)
		case CodeUnauthorized:
			return status.Error(codes.Unauthenticated, domainErr.Msg)
		}
	}

	return status.Error(codes.Internal, "internal error")
}
