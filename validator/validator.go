// Package validator provides convenient utilities for validation.
package validator

import (
	"context"

	validate "github.com/go-playground/validator/v10"
)

type (
	// Validator is a validation helper.
	Validator struct {
		v   *validate.Validate
		tr  func(ctx context.Context, errs validate.ValidationErrors) error
		tag string
	}
	// Option is a functional option for Validator.
	Option func(v *Validator)
)

// ErrorTranslator sets the error translator function for the validator.
// The translator will be used to translate validation errors into a custom error type.
func ErrorTranslator(t func(ctx context.Context, errs validate.ValidationErrors) error) Option {
	return func(v *Validator) {
		v.tr = t
	}
}

// New return new instance of validator with the given tag.
func New(tag string, opts ...Option) *Validator {
	validator := &Validator{
		v:   validate.New(),
		tag: tag,
	}
	if tag != "" {
		validator.v.SetTagName(tag)
	}

	for _, opt := range opts {
		opt(validator)
	}
	return validator
}

// Validate a struct exposed fields base on the definition of validate tag.
func (validator *Validator) Validate(ctx context.Context, v any) error {
	return validator.translateErr(ctx, validator.v.StructCtx(ctx, v))
}

// ValidatePartial validates the fields passed in only, ignoring all others.
func (validator *Validator) ValidatePartial(ctx context.Context, v any, fields ...string) error {
	return validator.translateErr(ctx, validator.v.StructPartialCtx(ctx, v, fields...))
}

// ValidateExcept validates all the fields except the given fields.
func (validator *Validator) ValidateExcept(ctx context.Context, v any, fields ...string) error {
	return validator.translateErr(ctx, validator.v.StructExceptCtx(ctx, v, fields...))
}

// Var validates a single variable using tag style validation.
func (validator *Validator) Var(ctx context.Context, field any, tag string) error {
	return validator.translateErr(ctx, validator.v.VarCtx(ctx, field, tag))
}

// Register adds a validation with the given tag
func (validator *Validator) Register(tag string, fn validate.FuncCtx, callValidationEvenIfNull bool) error {
	return validator.v.RegisterValidationCtx(tag, fn, callValidationEvenIfNull)
}

// translateErr translates the validation errors using the translator function if set.
func (validator *Validator) translateErr(ctx context.Context, err error) error {
	if err == nil || validator.tr == nil {
		return err
	}
	if errs, ok := err.(validate.ValidationErrors); ok && len(errs) > 0 {
		return validator.tr(ctx, errs)
	}
	return err
}
