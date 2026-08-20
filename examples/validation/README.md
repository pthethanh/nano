# gRPC request validation

This example shows nano's complete request-validation path without unrelated
middleware or application infrastructure:

1. `proto/validation.proto` declares email and age rules with `buf.validate`.
2. The server installs nano's `validator.UnaryServerInterceptor`.
3. The client sends one valid request and one invalid request.
4. The invalid request is rejected before the handler with
   `codes.InvalidArgument` and structured violation details.

Start the server:

```shell
go run ./server
```

In another terminal, run the client:

```shell
go run ./client
```

The client prints output similar to:

```text
valid request: registered developer@example.com
invalid request: code=InvalidArgument message="validation error: ..."
  field=email rule=string.email message="must be a valid email address"
  field=age rule=uint32.gte_lte message="must be greater than or equal to 18 and less than or equal to 120"
```

To regenerate the protobuf files after changing the schema:

```shell
make install_tools
make gen_proto
```

This module is part of the repository's `go.work`; use `go work sync` from the
repository root when dependency wiring changes. Its nano requirement remains on
the intended released version while local development resolves the workspace
copy.
