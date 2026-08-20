package validator

import (
	"buf.build/go/protovalidate"
	"google.golang.org/protobuf/proto"
)

type (
	// Validator validates protobuf messages using buf.validate rules declared
	// in their protobuf schemas.
	Validator = protovalidate.Validator

	// Option configures a Validator created by New.
	Option = protovalidate.ValidatorOption

	// ValidationOption configures an individual validation call.
	ValidationOption = protovalidate.ValidationOption
)

// New creates a protobuf validator.
func New(opts ...Option) (Validator, error) {
	return protovalidate.New(opts...)
}

// Default returns the shared, concurrency-safe protobuf validator.
func Default() Validator {
	return protovalidate.GlobalValidator
}

// Validate validates msg using the shared validator.
func Validate(msg proto.Message, opts ...ValidationOption) error {
	return protovalidate.Validate(msg, opts...)
}
