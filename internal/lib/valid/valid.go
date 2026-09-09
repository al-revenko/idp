package valid

import (
	"errors"
	"strings"

	"github.com/al-revenko/idp/internal/lib/sign"
	"github.com/go-playground/validator/v10"
)

var pkg = sign.Pkg("lib/valid")

type Valid struct {
	validate *validator.Validate
}

func New() *Valid {
	v := validator.New()

	v.RegisterValidation("password", tagPassword)

	return &Valid{validate: v}
}

func (v *Valid) Struct(s any) error {
	op := pkg.Op("Struct")

	err := v.validate.Struct(s)
	if err == nil {
		return nil
	}

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		return op.Err(ValidationError{cause: ve})
	}

	return op.Err(ValidationError{cause: err})
}

func tagPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	var hasLower, hasUpper, hasDigit, hasSpecial bool

	for _, char := range password {
		switch {
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case strings.ContainsRune("@$!%*?&", char):
			hasSpecial = true
		default:
			return false
		}
	}

	return hasLower && hasUpper && hasDigit && hasSpecial
}
