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

## Package Boundaries
- Treat top-level packages as independent modules inside the repository. A top-level package must not import another top-level package from this repository.
- The same rule applies to subpackages. For example, code under `grpc/...` must not import `validator`, `metric`, `log`, `status`, `cache`, `broker`, `config`, or other top-level packages unless the exception is explicitly documented here.
- If functionality from another top-level package is needed, define a small local interface in the consuming package and inject the real implementation at runtime.
- Prefer constructor injection or option-based injection over package-level reach-through.
- Tests should follow the same rule by default. Prefer local fakes, stubs, or adapters instead of importing a real implementation from another top-level package.
- Before finalizing any change, inspect new imports in touched files. If a new import crosses a top-level package boundary, stop and redesign around an interface unless an explicit exception exists in this file.

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
5. Verify that touched files did not introduce cross-package dependencies between top-level packages. If they did, redesign around a local interface and runtime injection.
6. Summarize changed files, behavior, and any follow-up work.
