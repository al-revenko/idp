package domain

import "errors"

type DomainErrCode uint8

const (
	CodeInvalidInput DomainErrCode = iota
	CodeNotFound
	CodeConflict
	CodeForbidden
	CodeUnauthenticated
	CodeInternal
)

type DomainError struct {
	Code  DomainErrCode
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

func Error(code DomainErrCode, msg string, cause error) DomainError {
	return DomainError{Code: code, Msg: msg, Cause: cause}
}

func IsErrCode(err error, code DomainErrCode) bool {
	if e, ok := errors.AsType[DomainError](err); ok {
		return e.Code == code
	}

	return false
}
