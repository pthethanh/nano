package validator_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	validate "github.com/go-playground/validator/v10"
	"github.com/pthethanh/nano/validator"
)

func TestValidatePartial(t *testing.T) {
	type Note struct {
		Value string `validate:"required"`
	}
	type Address struct {
		Work string `validate:"required"`
		Home string `validate:"required"`
		Note Note
	}
	type Employee struct {
		Name     string `validate_me:"required"`
		Age      int    `validate:"gt=1"`
		Address1 Address
		Note     string `validate:"len=10"`
	}
	cases := []struct {
		name     string
		tag      string
		value    any
		fields   []string
		excepts  []string
		field    bool
		fieldTag string
		err      bool
	}{
		{
			name: "some fields, validate success",
			value: Employee{
				Name: "",
				Age:  2,
			},
			fields: []string{"Name", "Age"},
			err:    false,
		},
		{
			name: "nested field, required tag not provide value -> failed",
			value: Employee{
				Name: "",
				Age:  2,
			},
			fields: []string{"Address1.Note.Value"},
			err:    true,
		},
		{
			name: "except fields, validate success",
			value: Employee{
				Note: "1234567890",
			},
			excepts: []string{"Name", "Age", "Address1"},
			err:     false,
		},
		{
			name:     "email validation, validate failed",
			fieldTag: "email",
			field:    true,
			value:    "test@",
			err:      true,
		},
		{
			name:     "email validation, validate success",
			fieldTag: "email",
			field:    true,
			value:    "test@gmail.com",
			err:      false,
		},
		{
			name: "custom tag, validate fail",
			value: Employee{
				Name: "",
			},
			tag: "validate_me",
			err: true,
		},
		{
			name: "custom tag, validate success",
			value: Employee{
				Name: "hello",
			},
			tag: "validate_me",
			err: false,
		},
	}
	if root := validator.Default(); root == nil {
		t.Error("root should not return nil")
	}
	tr := validator.ErrorTranslator(func(ctx context.Context, errs validate.ValidationErrors) error {
		return fmt.Errorf("nano_validation_error: %w", errs)
	})
	validator.SetDefault(validator.New("", tr))
	validator.SetDefault(validator.New("validate_me", tr))
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var err error
			if c.fields != nil {
				err = validator.Default(c.tag).ValidatePartial(context.Background(), c.value, c.fields...)
			} else if c.excepts != nil {
				err = validator.Default(c.tag).ValidateExcept(context.Background(), c.value, c.excepts...)
			} else if c.field {
				err = validator.Default("").Var(context.Background(), c.value, c.fieldTag)
			} else {
				err = validator.Default(c.tag).Validate(context.Background(), c.value)
			}
			if c.err && err != nil && !strings.Contains(err.Error(), "nano_validation_error") {
				t.Errorf("got err=%v, want nano_validation_error", err)
			}
			if c.err && err == nil {
				t.Errorf("got validation success, want validation fail.")
			}
			if !c.err && err != nil {
				t.Errorf("got validation err=%v, want validation success.", err)
			}
		})
	}
}
