package helpers

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	// Use json tag names in error messages so they match what the client sends.
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})
}

// ValidateStruct runs validation on s and returns a human-readable error string
// listing every failing field. Returns nil when validation passes.
func ValidateStruct(s interface{}) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var msgs []string
	for _, e := range err.(validator.ValidationErrors) {
		field := e.Field()
		switch e.Tag() {
		case "required":
			msgs = append(msgs, fmt.Sprintf("%s is required", field))
		case "email":
			msgs = append(msgs, fmt.Sprintf("%s must be a valid email address", field))
		case "oneof":
			msgs = append(msgs, fmt.Sprintf("%s must be one of %s", field, e.Param()))
		case "uuid":
			msgs = append(msgs, fmt.Sprintf("%s must be a valid UUID", field))
		case "gt":
			msgs = append(msgs, fmt.Sprintf("%s must be greater than %s", field, e.Param()))
		case "gte":
			msgs = append(msgs, fmt.Sprintf("%s must be at least %s", field, e.Param()))
		case "min":
			msgs = append(msgs, fmt.Sprintf("%s must be at least %s characters", field, e.Param()))
		case "max":
			msgs = append(msgs, fmt.Sprintf("%s must be at most %s characters", field, e.Param()))
		default:
			msgs = append(msgs, fmt.Sprintf("%s is invalid", field))
		}
	}
	return fmt.Errorf("%s", strings.Join(msgs, "; "))
}
