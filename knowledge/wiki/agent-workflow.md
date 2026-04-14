# Agent Workflow

Use this order by default:
1. Read the root `AGENTS.md`.
2. Read [`knowledge/index.md`](../index.md).
3. Read the most relevant pages under `knowledge/wiki/`.
4. Read nearby package files and tests.
5. Make the smallest change that solves the task.
6. Run the narrowest validation command that proves the change.
7. Run `./scripts/check-boundaries.sh` when touching imports or package structure.
8. Update the knowledge base for durable findings or instructions.
9. Append a short entry to [`knowledge/log.md`](../log.md).

Knowledge-base maintenance rules:
- update existing pages instead of creating duplicates when possible
- keep pages concise and cross-linked
- keep `raw/` notes immutable unless explicitly refreshing them
- treat `index.md` and `log.md` as required maintenance surfaces, not optional docs
