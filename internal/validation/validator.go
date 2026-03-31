package validation

import (
	"context"

	errs "github.com/reststore/restkit/internal/errors"
)

var DefaultValidator func(ctx context.Context, s any) errs.ValidationResult

func ValidateStruct(ctx context.Context, s any) errs.ValidationResult {
	if DefaultValidator != nil {
		return DefaultValidator(ctx, s)
	}
	return errs.ValidationResult{}
}
