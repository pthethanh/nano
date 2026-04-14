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
- Agents should start with this file, then read `knowledge/index.md` and the most relevant `knowledge/wiki/*.md` pages before broad exploration.

## Knowledge Base
- This repository maintains a persistent agent-written knowledge base under `knowledge/`.
- Structure:
  - `knowledge/raw/`: immutable source notes and curated external references. Agents may add files here but should not rewrite existing raw source captures unless explicitly asked to refresh them.
  - `knowledge/index.md`: content-oriented index of the knowledge base. Update it whenever adding or renaming a knowledge page.
  - `knowledge/log.md`: append-only chronological log of important ingests, decisions, and maintenance passes. Add a short entry for every substantial task.
  - `knowledge/wiki/`: synthesized markdown pages that summarize architecture, package behavior, decisions, and recurring agent instructions.
- Before doing broad repo exploration, read `knowledge/index.md` and the most relevant `knowledge/wiki/*.md` pages first. Use the knowledge base to avoid rediscovering stable repo context from scratch.
- When you learn something durable about the repository, package behavior, workflow, or architecture, update the relevant page in `knowledge/wiki/` instead of leaving it only in chat history.
- When the user gives a durable instruction about workflow, repository rules, file layout, or preferred patterns, update both `AGENTS.md` and the relevant page under `knowledge/wiki/`, then append a note to `knowledge/log.md`.
- Keep knowledge pages concise, factual, and cross-linked. Prefer updating an existing page over creating duplicates.
- Before finalizing any non-trivial task, either update the knowledge base for durable findings or explicitly confirm in the final response why no knowledge update was needed.
- Keep workflow guidance centralized here and in `knowledge/wiki/` instead of duplicating it across package-local `AGENTS.md` files.
- Do not bump submodule `require github.com/pthethanh/nano vX.Y.Z` selectors just to make local wiring work. Keep the selector on the intended released version.
- For modules included in `go.work`, prefer workspace resolution over local `replace github.com/pthethanh/nano ...` directives.

## Task Routing
- Interceptor work:
  - read `knowledge/wiki/interceptor-notes.md`
  - prefer option-based construction and safe defaults
  - keep interface or interceptor wiring in the entry file and concrete implementations in separate files
  - validate with `go test ./grpc/interceptor/...`
- gRPC client work:
  - keep helpers thin and aligned with grpc-go and Kubernetes-first usage
  - prefer native grpc-go dial options and config types
  - validate with `go test ./grpc/client`
- gRPC server work:
  - preserve grpc-go server semantics
  - keep lifecycle and option behavior explicit
  - validate with `go test ./grpc/server`
- Generator work:
  - change generator logic before generated outputs
  - regenerate affected outputs after generator changes
  - validate with `go test ./cmd/protoc-gen-nano/...`

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
- When a package exposes an interface plus one or more default or concrete implementations, keep the interface/interceptor entrypoints in one file and put each concrete implementation in its own file in the same package. For example, keep interceptor wiring in `*.go` and move concrete implementations such as token buckets, semaphores, or threshold breakers into `token_bucket.go`, `semaphore.go`, or `threshold.go`.
- When retry behavior needs backoff, prefer `google.golang.org/grpc/backoff.Config` semantics over introducing custom backoff APIs.

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
- `knowledge/wiki/*.md`: centralized persistent repo knowledge and workflow guidance

## Avoid
- broad repo-wide refactors without a clear need
- editing unrelated modules while solving a local issue
- introducing framework-like abstractions for simple helper logic
- changing generated files without updating the generator or source input

## Preferred Agent Workflow
1. Read `AGENTS.md`, `knowledge/index.md`, and the most relevant `knowledge/wiki/*.md` pages before broad code exploration.
2. Read the nearest package files and tests before editing.
3. Make the smallest change that satisfies the task.
4. Update or add tests when behavior changes.
5. Run focused validation commands.
6. Run `./scripts/check-boundaries.sh` when touching imports, package structure, or public package layout.
7. Verify that touched files did not introduce cross-package dependencies between top-level packages. If they did, redesign around a local interface and runtime injection.
8. Update the knowledge base for durable findings or new standing instructions, then append a short entry to `knowledge/log.md`.
9. Summarize changed files, behavior, and any follow-up work.
