package valid

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/iancoleman/strcase"
)

type ValidationError struct {
	msg   string
	cause error
}

func (e *ValidationError) Error() string {
	var ve validator.ValidationErrors

	if e.cause == nil {

	}

	if errors.As(e.cause, &ve) {
		var builder strings.Builder

		builder.Grow(len(ve))

		for i, fe := range ve {
			field := strcase.ToSnake(fe.Field())
			endOfmsg := "; "

			if len(ve) == 1 || i == len(ve)-1 {
				endOfmsg = ""
			}

			if fe.Tag() == "required" {
				builder.WriteString(fmt.Sprintf("%s: %s%s", field, fe.Tag(), endOfmsg))
			} else if fe.Param() != "" {
				builder.WriteString(fmt.Sprintf("%s: should be %s=%s%s", field, fe.Tag(), fe.Param(), endOfmsg))
			} else {
				builder.WriteString(fmt.Sprintf("%s: should be %s%s", field, fe.Tag(), endOfmsg))
			}
		}

		return builder.String()
	}

	if e.msg != "" {
		return e.msg
	}

	return "validation failed"
}

func (e *ValidationError) Unwrap() error {
	return e.cause
}
