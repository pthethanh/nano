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

## Concurrency tests: testing/synctest
- when a test only synchronizes in-process goroutines with `time.Sleep` (timers, TTLs, polling loops), prefer wrapping it in `testing/synctest.Test` with `synctest.Wait()` instead of real sleeps; see `cache/memory/memory_test.go` `TestCacheTimeout` for the pattern
- do not use `testing/synctest` for tests that hit real network sockets or external processes (real HTTP servers, external brokers like Kafka) — synctest requires every bubble goroutine to be "durably blocked" on in-bubble primitives, and real I/O blocking does not qualify; it will deadlock/panic
- wrapping a test in synctest can surface pre-existing goroutine leaks (an unclosed background worker keeps the bubble's root goroutine from exiting cleanly) — treat that as a real bug to fix in the test, not a reason to avoid synctest
