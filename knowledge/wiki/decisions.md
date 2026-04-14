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

## Dependencies
- Keep dependencies minimal.
- Avoid bringing in large or test-only dependencies when local fakes or API-level tests are sufficient.
