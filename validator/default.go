package validator

import (
	"context"
	"sync"

	validate "github.com/go-playground/validator/v10"
)

var (
	cache sync.Map
)

// Get get the tag validator from cache or create new one if not exists.
// If tag is empty, return Root() validator.
func Get(tag string) *Validator {
	if tag == "" {
		return Root()
	}
	if v, ok := cache.Load(tag); ok {
		return v.(*Validator)
	}
	v := New(tag)
	cache.Store(tag, v)
	return v
}

// Validate a struct exposed fields base on the definition of validate tag.
func Validate(ctx context.Context, v any) error {
	return Root().Validate(ctx, v)
}

// ValidatePartial validates the fields passed in only, ignoring all others.
func ValidatePartial(ctx context.Context, v any, fields ...string) error {
	return Root().ValidatePartial(ctx, v, fields...)
}

// ValidateExcept validates all the fields except the given fields.
func ValidateExcept(ctx context.Context, v any, fields ...string) error {
	return Root().ValidateExcept(ctx, v, fields...)
}

// Var validates a single variable using tag style validation.
func Var(ctx context.Context, field any, tag string) error {
	return Root().Var(ctx, field, tag)
}

// Register adds a validation with the given tag
func Register(tag string, fn validate.FuncCtx, callValidationEvenIfNull bool) error {
	return Root().Register(tag, fn, callValidationEvenIfNull)
}
