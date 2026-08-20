package validator_test

import (
	"context"
	"errors"
	"testing"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/pthethanh/nano/validator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

func TestValidateUsesRulesDeclaredInProtobufSchema(t *testing.T) {
	messageType := requiredNameMessageType(t)
	message := messageType.New()

	if err := validator.Validate(message.Interface()); err == nil {
		t.Fatal("Validate() accepted an empty name that violates min_len=1")
	}

	name := message.Descriptor().Fields().ByName("name")
	message.Set(name, protoreflect.ValueOfString("nano"))
	if err := validator.Validate(message.Interface()); err != nil {
		t.Fatalf("Validate() rejected a valid message: %v", err)
	}
}

func TestUnaryServerInterceptorRejectsInvalidProtobufWithDetails(t *testing.T) {
	request := requiredNameMessageType(t).New().Interface()
	called := false
	interceptor := validator.UnaryServerInterceptor(validator.Default())

	_, err := interceptor(context.Background(), request, &grpc.UnaryServerInfo{}, func(context.Context, any) (any, error) {
		called = true
		return nil, nil
	})
	if called {
		t.Fatal("handler was called with an invalid request")
	}
	if got := status.Code(err); got != codes.InvalidArgument {
		t.Fatalf("status code = %v, want %v", got, codes.InvalidArgument)
	}
	details := status.Convert(err).Details()
	if got := len(details); got != 1 {
		t.Fatalf("status details = %d, want one structured violation detail", got)
	}
	if _, ok := details[0].(*validate.Violations); !ok {
		t.Fatalf("status detail type = %T, want *validate.Violations", details[0])
	}
}

func TestStreamServerInterceptorRejectsInvalidReceivedMessage(t *testing.T) {
	messageType := requiredNameMessageType(t)
	interceptor := validator.StreamServerInterceptor(nil)
	stream := &fakeServerStream{}

	err := interceptor(nil, stream, &grpc.StreamServerInfo{}, func(_ any, stream grpc.ServerStream) error {
		return stream.RecvMsg(messageType.New().Interface())
	})
	if got := status.Code(err); got != codes.InvalidArgument {
		t.Fatalf("status code = %v, want %v", got, codes.InvalidArgument)
	}
}

func TestUnaryServerInterceptorRejectsNonProtobufRequestAsInternal(t *testing.T) {
	interceptor := validator.UnaryServerInterceptor(nil)
	_, err := interceptor(context.Background(), "not protobuf", &grpc.UnaryServerInfo{}, func(context.Context, any) (any, error) {
		t.Fatal("handler was called with a non-protobuf request")
		return nil, nil
	})
	if got := status.Code(err); got != codes.Internal {
		t.Fatalf("status code = %v, want %v", got, codes.Internal)
	}
}

func TestUnaryServerInterceptorReturnsValidatorFailureAsInternal(t *testing.T) {
	interceptor := validator.UnaryServerInterceptor(validatorFunc(func(proto.Message, ...validator.ValidationOption) error {
		return errors.New("invalid validation rule")
	}))
	_, err := interceptor(context.Background(), requiredNameMessageType(t).New().Interface(), &grpc.UnaryServerInfo{}, func(context.Context, any) (any, error) {
		t.Fatal("handler was called after the validator failed")
		return nil, nil
	})
	if got := status.Code(err); got != codes.Internal {
		t.Fatalf("status code = %v, want %v", got, codes.Internal)
	}
}

type validatorFunc func(proto.Message, ...validator.ValidationOption) error

func (fn validatorFunc) Validate(message proto.Message, opts ...validator.ValidationOption) error {
	return fn(message, opts...)
}

type fakeServerStream struct {
	grpc.ServerStream
}

func (*fakeServerStream) Context() context.Context { return context.Background() }
func (*fakeServerStream) RecvMsg(any) error        { return nil }

func requiredNameMessageType(t *testing.T) protoreflect.MessageType {
	t.Helper()

	fieldOptions := &descriptorpb.FieldOptions{}
	proto.SetExtension(fieldOptions, validate.E_Field, validate.FieldRules_builder{
		String: validate.StringRules_builder{MinLen: proto.Uint64(1)}.Build(),
	}.Build())

	file := &descriptorpb.FileDescriptorProto{
		Name:       proto.String("nano.validator.request.proto"),
		Package:    proto.String("nano.validator.test"),
		Syntax:     proto.String("proto3"),
		Dependency: []string{"buf/validate/validate.proto"},
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: proto.String("Request"),
			Field: []*descriptorpb.FieldDescriptorProto{{
				Name:    proto.String("name"),
				Number:  proto.Int32(1),
				Type:    descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
				Label:   descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				Options: fieldOptions,
			}},
		}},
	}

	files := new(protoregistry.Files)
	if err := files.RegisterFile(validate.File_buf_validate_validate_proto); err != nil {
		t.Fatalf("register validate descriptor: %v", err)
	}
	descriptor, err := protodesc.FileOptions{}.New(file, files)
	if err != nil {
		t.Fatalf("build request descriptor: %v", err)
	}
	return dynamicpb.NewMessageType(descriptor.Messages().ByName("Request"))
}
