# Knowledge Log

## [2026-04-14] ingest | repository knowledge base bootstrap
- Created `knowledge/` with `raw/`, `wiki/`, `index.md`, and `log.md`.
- Added persistent agent workflow rules to `AGENTS.md` and `CLAUDE.md`.
- Seeded the wiki with repository overview, architecture, workflow, standing instructions, and interceptor notes.

## [2026-04-14] ingest | standing instruction persistence
- Added a standing rule that durable user workflow and layout instructions must be written to both `AGENTS.md` and the knowledge base automatically.
- Added a standing rule that interface/interceptor entrypoints stay separate from concrete implementations in the same package.

## [2026-04-14] ingest | grpc-native retry backoff preference
- Recorded a standing preference that retry backoff should use `google.golang.org/grpc/backoff.Config` semantics instead of a custom exponential backoff API.
- Updated interceptor notes to keep future retry work aligned with native gRPC behavior.

## [2026-04-14] maintenance | stream interceptor semantics and test dependency trim
- Updated client stream tracing and metrics guidance to require lifecycle semantics instead of setup-only instrumentation under generic stream interceptor names.
- Recorded a standing preference to avoid root-module tracing SDK dependencies when API-level local test doubles are sufficient.
