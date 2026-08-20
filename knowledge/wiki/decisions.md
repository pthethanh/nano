# Decisions

Durable architectural decisions in this repository:

## Core direction
- `nano` is a modular toolkit, not a service framework.
- The primary deployment target is Kubernetes.
- Prefer native Go and grpc-go patterns over framework abstractions.

## Package structure
- Top-level packages stay independent.
- Cross-package behavior should use local interfaces and runtime injection instead of direct imports.
- Interfaces and interceptor entrypoints stay separate from concrete implementations.
- Agent workflow instructions are centralized in the root `AGENTS.md` and `knowledge/wiki/`, not duplicated in package-local `AGENTS.md` files.

## gRPC behavior
- Keep gRPC client and server helpers thin.
- `registry` is not a core package for the Kubernetes-first design.
- Retry behavior must be explicit and safe by default.
- Retry backoff should use `google.golang.org/grpc/backoff.Config` semantics.
- Stream tracing and stream metrics should reflect stream lifecycle, not only stream creation.
- In `grpc/server`, outer HTTP mounts must be applied from the most specific prefix to the least specific prefix, and duplicate top-level prefixes should fail fast instead of silently shadowing routes.
- `grpc/server.APIPrefix` is a gateway mount prefix, so the HTTP server should strip that prefix before forwarding requests into `grpc-gateway`.
- `grpc/server.AutoMaxProcs()` is opt-in (not default) container-aware `GOMAXPROCS` via stdlib `runtime.SetDefaultGOMAXPROCS()`, applied in `listenAndServe`. It is opt-in because it overrides an explicit `GOMAXPROCS` env var if the operator set one; enabling it by default would silently change that behavior.
- Lazy package-level singletons (`grpc/server.Default()`/`SetDefault()`, `log.Default()`/`SetDefault()`) use a bare `atomic.Pointer[T]` with `CompareAndSwap(nil, constructed)` in the getter — not `sync.Once` plus a separate Load/Store. `sync.Once` wrapped around a Load-then-Store is a check-then-act race against a concurrent `SetDefault`, and (for `log`) also means `SetDefault(nil)` permanently breaks the getter since `once` never re-fires. Follow the `CompareAndSwap` pattern for any new lazy-singleton-with-override getter.
- Sentinel errors returned by interceptor packages (`circuitbreaker.ErrOpen`, `ratelimit.ErrLimited`, `auth`'s `Err*` vars) must be constructed with `status.Error(codes.X, ...)`, not `errors.New`/`fmt.Errorf`. A plain error returned across a gRPC boundary surfaces to the client as `codes.Unknown`, breaking any code-based client logic (retry policies, `errors.Is` against codes). Sentinels stay comparable via `errors.Is`/`==` either way since the check is by identity, not by underlying type.
- Verifying a `cmd/protoc-gen-nano` generator change is inert (formatting/dedup-only) is done by rebuilding+reinstalling the plugin, regenerating `examples/helloworld` via `make gen_proto`, and diffing: the `.pb.nano.go` output should be unaffected. Unrelated diffs in `.pb.go`/`_grpc.pb.go` from newer locally-installed `protoc-gen-go`/`protoc-gen-go-grpc` binary versions are environment drift, not a generator behavior change — revert those before committing.

## Dependencies
- Keep dependencies minimal.
- Avoid bringing in large or test-only dependencies when local fakes or API-level tests are sufficient.
- `sarama.NewMockBroker` (already a transitive test dependency via `IBM/sarama`) gives an in-process fake Kafka broker for tests that need a real `sarama.NewClient`/producer lifecycle (e.g. proving a race in `Open()`), without needing a live Kafka cluster or a new dependency. No equivalent lightweight in-process fake exists for `nats.go` in this repo yet; NATS-specific behavior that needs a real `*nats.Conn` is either tested via pure helper-function extraction or left to the existing `NATS_TEST`/`KAFKA_TEST`-gated integration tests.

## Toolchain
- Minimum/target Go version is `go 1.27.0`, set in the root `go.mod`, `go.work`, and every submodule `go.mod` (`cmd/protoc-gen-nano`, `examples/*`, `plugins/*`). Keep these in sync when bumping the Go version.
- Generics are the preferred tool for type-safe reusable code (see `config.Reader[T]`, `cache/memory.Cacher[K,V]`, `grpc/interceptor/authz.FromAnyContext[T]`, `broker.Codec[T]`) over `any`/type-assertion helpers, matching the repo's existing style.
- `broker.Codec` (`broker/broker.go`) is generic: `Codec[T any] { Marshal(*T) ([]byte, error); Unmarshal([]byte, *T) error }`, mirroring `plugins/cache/redis.Codec[V]`. Every `Broker[T]` implementation (kafka, nats, watermill) stores `codec broker.Codec[T]` and defaults to a generic `JSONCodec[T]{}`/`jsonCodec[T]{}`. When adding a new broker plugin, follow this pattern instead of an `any`-typed codec.
- `grpc/interceptor/authz` context keys that support arbitrary caller types (`anyContextKey[T]`, used by `NewAnyContext`/`FromAnyContext`) must be generic types themselves (`type anyContextKey[T any] struct{}`), not a single shared non-generic key — a shared key silently collides across different T instantiations (last write wins across all types), defeating the "isolated per type" contract. `RequestFromContext`/`NewRequestContext` are generic too, returning `(T, bool)`.
- `broker.PublishOptions.Headers map[string]string` (set via `broker.Header`/`broker.Headers`) is the cross-transport per-publish metadata mechanism. Every network `Broker[T]` implementation (kafka, nats, watermill) must wire it into its wire format (kafka: `sarama.ProducerMessage.Headers`; nats: `nats.Header` via `conn.PublishMsg`, not `conn.Publish`; watermill: `message.Message.Metadata`) using a small pure `xHeadersFrom(map[string]string) X` helper kept unit-testable without I/O. `broker/memory` legitimately ignores it (no wire concept) — that's fine, not a gap to fix.
- Broker/cache `Address` options across plugins take variadic `...string` (`kafka.Address`, `nats.Address`, `plugins/cache/redis.Address`), never `[]string` or a single comma-joined `string`. `nats.Nats[T].addrs` is stored as `[]string` and joined with `,` only at the `nats.Connect` call site, since that's the one place the underlying client actually wants a comma-joined string.
