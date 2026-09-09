package valid

import (
	"errors"

	"github.com/al-revenko/idp/internal/lib/pkgmark"
	"github.com/go-playground/validator/v10"
)

var pkg = pkgmark.New("lib/valid")

var validate = validator.New()

func Struct(s any) error {
	op := pkg.Op("Struct")

	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		return op.Err(&ValidationError{cause: ve})
	}

	return op.Err(&ValidationError{cause: err})
}
