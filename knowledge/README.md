# Knowledge Base

This directory is the repo-local persistent knowledge base for `nano`.

It follows a simple LLM-maintained wiki pattern:
- `raw/`: immutable source notes and curated external references
- `index.md`: content index for the wiki
- `log.md`: append-only maintenance log
- `wiki/`: synthesized pages about architecture, packages, workflow, and agent instructions

Agents should read the index and the most relevant wiki pages before broad code exploration, then write durable findings back into the wiki as part of normal work.
