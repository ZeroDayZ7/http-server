package middleware

import (
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func ValidateStruct(s any) map[string]string {
	errs := make(map[string]string)
	if err := validate.Struct(s); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			errs[e.Field()] = e.Tag()
		}
	}
	return errs
}
