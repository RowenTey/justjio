package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom tag name function to use json tag names in errors
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// GetValidator returns the singleton validator instance for custom validation registration
func GetValidator() *validator.Validate {
	return validate
}

// ValidateStruct validates a struct and returns user-friendly error messages
func ValidateStruct(s any) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}

	var errorMessages []string
	for _, e := range validationErrors {
		errorMessages = append(errorMessages, formatValidationError(e))
	}

	return fmt.Errorf("%s", strings.Join(errorMessages, "; "))
}

// formatValidationError formats a single validation error into a user-friendly message
func formatValidationError(e validator.FieldError) string {
	field := e.Field()

	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "min":
		if e.Type().Kind() == reflect.String {
			return fmt.Sprintf("%s must be at least %s characters long", field, e.Param())
		}
		return fmt.Sprintf("%s must be at least %s", field, e.Param())
	case "max":
		if e.Type().Kind() == reflect.String {
			return fmt.Sprintf("%s must be at most %s characters long", field, e.Param())
		}
		return fmt.Sprintf("%s must be at most %s", field, e.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", field, e.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, e.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, e.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, e.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, e.Param())
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", field)
	case "numeric":
		return fmt.Sprintf("%s must be a valid number", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of [%s]", field, e.Param())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "uuid4":
		return fmt.Sprintf("%s must be a valid UUID v4", field)
	case "datetime":
		return fmt.Sprintf("%s must be a valid datetime", field)
	default:
		return fmt.Sprintf("%s is invalid (failed validation: %s)", field, e.Tag())
	}
}
