package validator

import (
	"context"
	"errors"

	"buf.build/go/protovalidate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// UnaryServerInterceptor validates each unary request before invoking its
// handler. A nil Validator uses Default.
func UnaryServerInterceptor(v Validator) grpc.UnaryServerInterceptor {
	if v == nil {
		v = Default()
	}
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if err := validateRequest(req, v); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor validates each message received from a stream
// before returning it to the handler. A nil Validator uses Default.
func StreamServerInterceptor(v Validator) grpc.StreamServerInterceptor {
	if v == nil {
		v = Default()
	}
	return func(srv any, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, &validatingServerStream{
			ServerStream: stream,
			v:            v,
		})
	}
}

type validatingServerStream struct {
	grpc.ServerStream
	v Validator
}

func (stream *validatingServerStream) RecvMsg(message any) error {
	if err := stream.ServerStream.RecvMsg(message); err != nil {
		return err
	}
	return validateRequest(message, stream.v)
}

func validateRequest(value any, v Validator) error {
	message, ok := value.(proto.Message)
	if !ok {
		return status.Errorf(codes.Internal, "validator: unsupported message type %T", value)
	}

	err := v.Validate(message)
	if err == nil {
		return nil
	}

	var validationErr *protovalidate.ValidationError
	if !errors.As(err, &validationErr) {
		return status.Error(codes.Internal, err.Error())
	}
	st := status.New(codes.InvalidArgument, err.Error())
	withDetails, detailErr := st.WithDetails(validationErr.ToProto())
	if detailErr != nil {
		return st.Err()
	}
	return withDetails.Err()
}
