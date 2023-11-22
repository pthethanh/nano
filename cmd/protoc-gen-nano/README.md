# protoc-gen-nano

This is protobuf code generation for nano. We use protoc-gen-nano to reduce boilerplate code.

## Install

```
go install github.com/pthethanh/nano/cmd/protoc-gen-nano
```

Also required: 

- [protoc](https://github.com/google/protobuf)
- [protoc-gen-go](https://google.golang.org/protobuf)

## Usage

1. Define your proto file normally
2. Generate code with option "nano_out"
3. Add option `--nano_opt generate_gateway=true` if you want to generate the gateway registration
3. Register your service with nano server

## LICENSE

protoc-gen-nano is a liberal reuse of protoc-gen-go hence we maintain the original license 