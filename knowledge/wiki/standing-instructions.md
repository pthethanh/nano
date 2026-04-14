# Standing Instructions

These instructions should persist across future agent sessions.

## Workflow persistence
- When the user gives a durable repository, workflow, or file-layout instruction, update both `AGENTS.md` and the relevant page in `knowledge/wiki/`.
- Append a short note to `knowledge/log.md` when doing so.
- Before finalizing any non-trivial task, either update the knowledge base or explicitly state why no knowledge update was needed.
- Keep agent workflow guidance centralized in the root `AGENTS.md` and `knowledge/wiki/` instead of duplicating it in package-local instruction files.
- Do not invent or bump local submodule dependency versions just to satisfy workspace wiring. Keep the intended released version.
- For modules already included in `go.work`, prefer workspace resolution over local `replace github.com/pthethanh/nano ...` directives.

## File layout rule
- When a package exposes an interface plus one or more default or concrete implementations, keep the interface or interceptor entrypoints in one file and move each concrete implementation into its own file in the same package.
- Examples:
  - `ratelimit.go` + `token_bucket.go`
  - `concurrencylimit.go` + `semaphore.go`
  - `circuitbreaker.go` + `threshold.go`

## Repository style
- Prefer native Go and gRPC patterns over framework-like abstractions.
- Keep top-level packages independent.
- Keep APIs small and explicit.
- Favor safe defaults over clever defaults.
- Prefer gRPC-native configuration types such as `google.golang.org/grpc/backoff.Config` over custom equivalents when modeling gRPC behavior.
- For tracing and metrics interceptors around streaming RPCs, prefer full stream-lifecycle semantics over setup-only instrumentation unless the API name explicitly says setup or dial.

## Agent entrypoint
- Agents should start with the root `AGENTS.md`, then read `knowledge/index.md` and the most relevant wiki pages.
- Run `./scripts/check-boundaries.sh` when touching imports, package structure, or public package layout.
