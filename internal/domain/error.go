package domain

type DomainErrCode uint8

const (
	CodeInvalidInput DomainErrCode = iota
	CodeNotFound
	CodeConflict
	CodeForbidden
	CodeUnauthorized
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
