# AGENTS.md

## Purpose
This repository contains `github.com/pthethanh/nano`, a modular Go toolkit for building microservices. It provides reusable packages for gRPC, broker, cache, config, logging, metrics, status handling, validation, and code generation.

The design intent in this repository is:
- keep packages simple and independently usable
- prefer native Go and gRPC patterns over heavy abstractions
- no cross-package dependencies
- make additive, low-risk changes by default

## Repository Map
- `grpc/client/`, `grpc/server/`, `grpc/interceptor/`: gRPC helpers, server/client wiring, interceptors
- `broker/`, `cache/`, `config/`, `log/`, `metric/`, `status/`, `validator/`: standalone library packages
- `cmd/protoc-gen-nano/`: nano code generator
- `plugins/`: optional plugin implementations with their own `go.mod` files
- `examples/helloworld/`, `examples/kafka/`: runnable reference integrations, also separate modules

## Working Rules
- Preserve the repository's modular design. Do not introduce dependencies between top-level packages.
- Prefer additive changes over refactors. Avoid renaming exported symbols or changing public behavior without a clear requirement.
- Follow existing package boundaries. Put new logic in the closest existing package instead of creating broad shared abstractions.
- Generated files should not be edited manually unless the task is explicitly about generated output or generator templates.
- When changing public behavior, add or update tests in the same area.
- Keep dependencies minimal.

## Coding Conventions
- Use standard Go formatting.
- Accept `context.Context` as the first parameter where appropriate.
- Return errors with enough context to debug failures.
- Keep APIs small and explicit.
- Avoid global mutable state unless the package already uses it deliberately.
- Match the existing style in the touched package rather than imposing a new pattern.

## Validation
Use the narrowest command that proves the change.

Common root commands:
- `go test ./...`
- `make test`
- `make fmt`
- `make vet`

Notes:
- The root `Makefile` default target runs `mod_tidy`, `fmt`, `vet`, `test`, and builds plugins/examples.
- Some directories under `plugins/` and `examples/` are separate Go modules. If you change one of them, run validation in that module as well.

## Code Generation
Generator-related work lives in `cmd/protoc-gen-nano/`.

Relevant commands from the root `Makefile`:
- `make gen_proto`
- `make install_tools`

If a task changes `.proto` definitions, generator behavior, or generated example output:
- update the source `.proto` or generator code first
- regenerate affected outputs
- verify the generated code builds and tests pass

## Good Places To Learn Patterns
- `README.md`: project intent and basic usage
- `examples/helloworld/`: end-to-end gRPC and gateway flow
- `grpc/client/*` and `grpc/server/*`: preferred API style for gRPC helpers
- existing `*_test.go` files: expected behavior and edge cases

## Avoid
- broad repo-wide refactors without a clear need
- editing unrelated modules while solving a local issue
- introducing framework-like abstractions for simple helper logic
- changing generated files without updating the generator or source input

## Preferred Agent Workflow
1. Read the nearest package files and tests before editing.
2. Make the smallest change that satisfies the task.
3. Update or add tests when behavior changes.
4. Run focused validation commands.
5. Summarize changed files, behavior, and any follow-up work.
