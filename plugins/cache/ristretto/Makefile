PROJECT_NAME=ristretto
GO_BUILD_ENV=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on
GO_FILES=$(shell go list ./... | grep -v /vendor/)

.SILENT:

default: mod_tidy fmt vet test

vet:
	$(GO_BUILD_ENV) go vet $(GO_FILES)

fmt:
	$(GO_BUILD_ENV) go fmt $(GO_FILES)

test:
	$(GO_BUILD_ENV) CGO_ENABLED=1 go test $(GO_FILES) -race -cover -count=1

upgrade_deps:
	$(GO_BUILD_ENV) go get -u ./...

mod_tidy:
	$(GO_BUILD_ENV) go mod tidy
