# Karpathy LLM Wiki Note

Source:
- https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f

Summary:
- The core pattern is to maintain a persistent markdown wiki between the agent and raw sources instead of rediscovering knowledge from raw documents on every query.
- The wiki is agent-written and continuously updated as new sources arrive or new questions are answered.
- `index.md` is the content-oriented navigation layer; `log.md` is the chronological activity layer.
- A schema file such as `AGENTS.md` or `CLAUDE.md` defines the maintenance workflow and conventions so future sessions behave consistently.

Adaptation for this repository:
- `knowledge/raw/` holds immutable source notes and curated references.
- `knowledge/wiki/` holds synthesized repository knowledge.
- `AGENTS.md` and `CLAUDE.md` tell agents to read and maintain the knowledge base before re-exploring stable repository context.
