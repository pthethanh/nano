package main

import (
	"context"
	"fmt"
	"strings"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/pthethanh/nano/examples/validation/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func main() {
	conn, err := grpc.NewClient("localhost:8082",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := api.NewRegistrationServiceClient(conn)
	callValidRequest(client)
	callInvalidRequest(client)
}

func callValidRequest(client api.RegistrationServiceClient) {
	response, err := client.Register(context.Background(), &api.RegisterRequest{
		Email: "developer@example.com",
		Age:   30,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("valid request: %s\n", response.GetMessage())
}

func callInvalidRequest(client api.RegistrationServiceClient) {
	_, err := client.Register(context.Background(), &api.RegisterRequest{
		Email: "not-an-email",
		Age:   12,
	})
	if err == nil {
		panic("invalid request unexpectedly succeeded")
	}

	st := status.Convert(err)
	fmt.Printf("invalid request: code=%s message=%q\n", st.Code(), st.Message())
	if st.Code() != codes.InvalidArgument {
		return
	}
	for _, detail := range st.Details() {
		violations, ok := detail.(*validate.Violations)
		if !ok {
			continue
		}
		for _, violation := range violations.GetViolations() {
			fmt.Printf("  field=%s rule=%s message=%q\n",
				fieldPath(violation.GetField()), violation.GetRuleId(), violation.GetMessage())
		}
	}
}

func fieldPath(path *validate.FieldPath) string {
	fields := make([]string, 0, len(path.GetElements()))
	for _, element := range path.GetElements() {
		fields = append(fields, element.GetFieldName())
	}
	return strings.Join(fields, ".")
}
