package validation

import (
	"context"

	routectx "github.com/reststore/restkit/internal/context"
	errs "github.com/reststore/restkit/internal/errors"
)

var DefaultValidator func(ctx context.Context, s any) errs.ValidationResult

func Validate(ctx context.Context, s any) errs.ValidationResult {
	if v := ctx.Value(routectx.ValidatorCtxKey); v != nil {
		if validator, ok := v.(func(context.Context, any) errs.ValidationResult); ok {
			return validator(ctx, s)
		}
	}

	if DefaultValidator != nil {
		return DefaultValidator(ctx, s)
	}

	return errs.ValidationResult{}
}
