# Repo Overview

`nano` is a modular Go toolkit for building microservices with an explicit preference for native Go and gRPC patterns over framework-like abstractions.

Core intent:
- keep top-level packages independently usable
- avoid cross-package dependencies between top-level packages
- make additive, low-risk changes
- keep dependencies and abstractions small

Top-level package map:
- `grpc/`: gRPC server, client, health, and interceptors
- `broker/`, `cache/`, `config/`, `log/`, `metric/`, `status/`, `validator/`: standalone packages
- `cmd/protoc-gen-nano/`: generator
- `plugins/`: optional implementations as separate modules
- `examples/`: runnable examples as separate modules

Important repo documents:
- [`README.md`](../../README.md)
- [`AGENTS.md`](../../AGENTS.md)
- [`CLAUDE.md`](../../CLAUDE.md)
- [`knowledge/wiki/architecture.md`](./architecture.md)
