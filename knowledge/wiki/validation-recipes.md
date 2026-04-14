# Validation Recipes

Use the narrowest command that proves the change.

## Common areas
- interceptors: `go test ./grpc/interceptor/...`
- client helpers: `go test ./grpc/client`
- server helpers: `go test ./grpc/server`
- metrics: `go test ./metric/...`
- generator: `go test ./cmd/protoc-gen-nano/...`

## Generator changes
1. update generator logic first
2. regenerate affected outputs
3. test generator package and affected examples

## Example modules
- `examples/helloworld`: run validation inside that module if changed
- `examples/kafka`: run validation inside that module if changed

## Plugin modules
- `plugins/broker/kafka`: run validation inside that module if changed
- `plugins/broker/nats`: run validation inside that module if changed
- `plugins/broker/watermill`: run validation inside that module if changed
- `plugins/cache/redis`: run validation inside that module if changed
- plugin and example modules that are already included in `go.work` should validate in workspace mode and avoid redundant local `replace github.com/pthethanh/nano ...` directives

## Boundary checks
- run `./scripts/check-boundaries.sh` when touching imports, adding packages, or changing package structure

## Final validation rule
- prefer focused package tests before broad repo-wide runs
- use broad validation only when the change spans multiple packages or surfaces
