package validator

import (
	"context"
	"sync"

	validate "github.com/go-playground/validator/v10"
)

var (
	def = new(sync.Map)
)

// SetDefault registers a default validator for the specified tag.
// If a validator is already registered under the same tag, it will be replaced.
// Calling SetDefault multiple times with different tags is safe.
// Example:
//
//	SetDefault(New("validate"))
//	SetDefault(New("custom_tag"))
func SetDefault(v *Validator) {
	def.Store(v.Tag(), v)
}

// Default returns the default validator instance of the given tag, creating one if needed.
// If no tag is provided, it uses the empty string as the default tag.
func Default(tags ...string) *Validator {
	tag := ""
	if len(tags) > 0 {
		tag = tags[0]
	}
	if v, ok := def.Load(tag); ok {
		return v.(*Validator)
	}
	v := New(tag)
	def.Store(tag, v)
	return v
}

// Validate a struct exposed fields base on the definition of validate tag.
func Validate(ctx context.Context, v any) error {
	return Default().Validate(ctx, v)
}

// ValidatePartial validates the fields passed in only, ignoring all others.
func ValidatePartial(ctx context.Context, v any, fields ...string) error {
	return Default().ValidatePartial(ctx, v, fields...)
}

// ValidateExcept validates all the fields except the given fields.
func ValidateExcept(ctx context.Context, v any, fields ...string) error {
	return Default().ValidateExcept(ctx, v, fields...)
}

// Var validates a single variable using tag style validation.
func Var(ctx context.Context, field any, tag string) error {
	return Default().Var(ctx, field, tag)
}

// Register adds a validation with the given tag
func Register(tag string, fn validate.FuncCtx, callValidationEvenIfNull bool) error {
	return Default().Register(tag, fn, callValidationEvenIfNull)
}
