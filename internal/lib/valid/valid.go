package valid

import (
	"errors"
	"fmt"

	"github.com/al-revenko/idp/internal/lib/apperr"
	"github.com/al-revenko/idp/internal/lib/pkgmark"
	"github.com/go-playground/validator/v10"
	"github.com/iancoleman/strcase"
)

var pkg = pkgmark.New("lib/valid")

var validate = validator.New()

func RequestDTO(s any) error {
	op := pkg.Op("RequestDTO")

	if s == nil {
		return op.Err(apperr.New(apperr.CodeInternal, "nil passed instead of struct"))
	}

	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		ferrs := make([]error, len(ve))
		for i, fe := range ve {
			field := strcase.ToSnake(fe.Field())
			endOfmsg := "; "

			if len(ve) == 1 || i == len(ve)-1 {
				endOfmsg = ""
			}

			if fe.Tag() == "required" {
				ferrs[i] = fmt.Errorf("%s: %s%s", field, fe.Tag(), endOfmsg)
			} else if fe.Param() != "" {
				ferrs[i] = fmt.Errorf("%s: should be %s=%s%s", field, fe.Tag(), fe.Param(), endOfmsg)
			} else {
				ferrs[i] = fmt.Errorf("%s: should be %s%s", field, fe.Tag(), endOfmsg)
			}
		}

		out := errors.Join(ferrs...)

		return op.Err(apperr.From(apperr.CodeInvalidInput, out.Error(), out))
	}

	return op.Err(apperr.From(apperr.CodeInternal, err.Error(), err))
}
