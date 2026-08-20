GO_BUILD_ENV=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on
GO_FILES=$(shell go list ./... | grep -v /vendor/)

GOPATH ?= $(HOME)/go
PROTO_OUT = $(GOPATH)/src
PROTOVALIDATE := $(GOPATH)/src/github.com/bufbuild/protovalidate
PROTOVALIDATE_PROTO := $(PROTOVALIDATE)/proto/protovalidate
PROTOC_INCLUDES := /usr/local/include

export PATH := $(GOPATH)/bin:$(PATH)

.SILENT:

default: fmt vet test

vet:
	$(GO_BUILD_ENV) go vet $(GO_FILES)

fmt:
	$(GO_BUILD_ENV) go fmt $(GO_FILES)

test:
	$(GO_BUILD_ENV) CGO_ENABLED=1 go test $(GO_FILES) -race -cover -count=1

mod_sync:
	cd ../.. && go work sync

install_tools:
	go install tool
	git -C $(PROTOVALIDATE) pull || git clone https://github.com/bufbuild/protovalidate $(PROTOVALIDATE)

gen_proto:
	protoc -I $(PROTOC_INCLUDES) -I $(PROTOVALIDATE_PROTO) -I ./proto \
	 --go_out $(PROTO_OUT) \
	 --nano_out $(PROTO_OUT) \
	 --nano_opt generate_gateway=false \
	 --nano_opt generate_utils=false \
	 --go-grpc_out $(PROTO_OUT) \
	 ./proto/validation.proto
