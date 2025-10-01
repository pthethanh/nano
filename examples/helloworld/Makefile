PROJECT_NAME=helloword
GO_BUILD_ENV=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on
GO_FILES=$(shell go list ./... | grep -v /vendor/)

PROTOC_VERSION = 31.1
GOPATH ?= $(HOME)/go
PROTO_OUT = $(GOPATH)/src
MOD=$(GOPATH)/pkg/mod
GOOGLE_APIS_PROTO := $(GOPATH)/src/github.com/googleapis/googleapis
PROTOC_INCLUDES := /usr/local/include

export PATH := $(GOPATH)/bin:$(PATH)

.SILENT:

default: mod_tidy fmt vet test

vet:
	$(GO_BUILD_ENV) go vet $(GO_FILES)

fmt:
	$(GO_BUILD_ENV) go fmt $(GO_FILES)

test:
	$(GO_BUILD_ENV) CGO_ENABLED=1 go test $(GO_FILES) -race -cover -count=1

mod_tidy:
	$(GO_BUILD_ENV) go mod tidy

install_tools: install_protobuf
	go install tool
	git -C $(GOOGLE_APIS_PROTO) pull || git clone https://github.com/googleapis/googleapis $(GOOGLE_APIS_PROTO)

install_protobuf:
	wget https://github.com/google/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-linux-x86_64.zip
	unzip protoc-$(PROTOC_VERSION)-linux-x86_64.zip -d protoc
	sudo cp protoc/bin/protoc /usr/local/bin
	sudo mkdir -p /usr/local/include
	sudo cp -R protoc/include/* /usr/local/include/
	rm -rf protoc
	rm -rf protoc-$(PROTOC_VERSION)-linux-x86_64.zip
	sudo chmod -R 755 /usr/local/include/
	sudo chmod +x /usr/local/bin/protoc

gen_proto:
	$(PROTOC_ENV) protoc -I $(PROTOC_INCLUDES) -I $(GOOGLE_APIS_PROTO) -I ./proto \
	 --go_out $(PROTO_OUT) \
	 --nano_out $(PROTO_OUT) \
	 --nano_opt generate_gateway=true \
	 --nano_opt generate_utils=false \
	 --go-grpc_out $(PROTO_OUT) \
	 --grpc-gateway_out $(PROTO_OUT) \
     --grpc-gateway_opt logtostderr=true \
     --grpc-gateway_opt generate_unbound_methods=true \
	 --openapiv2_out docs \
     ./proto/helloworld.proto
