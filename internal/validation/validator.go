package validation

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	errs "github.com/reststore/restkit/internal/errors"
)

// validate is the instance of the validator used for struct validation
var validate = validator.New()

// ValidateStruct validates a struct using go-playground/validator tags
func ValidateStruct(ctx context.Context, s any) errs.ValidationResult {
	result := errs.ValidationResult{}

	// Handle nil input gracefully
	if s == nil {
		result.Status = 422
		result.Code = errs.ErrCodeValidation
		result.Message = errs.ErrMsgValidation
		result.Errors = append(result.Errors, errs.ValidationError{
			Field:   "",
			Message: "request body is required",
		})
		return result
	}

	if err := validate.Struct(s); err != nil {
		result.Status = 422
		result.Code = errs.ErrCodeValidation
		result.Message = errs.ErrMsgValidation

		if validationErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
			for _, e := range validationErrors {
				field := strings.ToLower(e.Field())
				message := getErrorMessage(e)
				result.Errors = append(result.Errors, errs.ValidationError{
					Field:   field,
					Message: message,
				})
			}
		}
	}

	return result
}

func getErrorMessage(e validator.FieldError) string {
	tag := e.Tag()
	field := e.Field()
	param := e.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s is not a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, param)
	case "len":
		return fmt.Sprintf("%s must be %s characters long", field, param)
	case "gte":
		return fmt.Sprintf(
			"%s must be greater than or equal to %s",
			field,
			param,
		)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	case "eq":
		return fmt.Sprintf("%s must be equal to %s", field, param)
	case "ne":
		return fmt.Sprintf("%s must not be equal to %s", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of %s", field, param)
	default:
		return fmt.Sprintf("%s failed validation on %s", field, tag)
	}
}
