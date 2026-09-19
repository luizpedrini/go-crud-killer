# Docs

User-facing design for **go-crud-killer**, produced by `/setup-matt-pocock-skills` then `/grill-with-docs` (`/grilling` + `/domain-modeling`).

Start here:

1. [Grill session](grilling/session.md) — design tree, settled Round 0, **Round 1 questions still waiting**, contingent Round 2.
2. [Proposed design](design.md) — recommended library shape if Round 1 recommendations are accepted. Not approved until you say so.
3. [Glossary](../CONTEXT.md) — settled domain language only.
4. [ADRs](adr/) — settled, hard-to-reverse decisions.

## Layout

| Path | What it is |
| --- | --- |
| [CONTEXT.md](../CONTEXT.md) | Glossary (domain-modeling). Implementation-free. |
| [docs/adr/0001-go-library-for-entity-crud.md](adr/0001-go-library-for-entity-crud.md) | Product is a Go library. |
| [docs/adr/0002-bitemporal-audit-on-every-entity.md](adr/0002-bitemporal-audit-on-every-entity.md) | Every Entity is bitemporal, auditable, append-only on transaction time. |
| [docs/grilling/session.md](grilling/session.md) | The interview. Open frontier lives here. |
| [docs/design.md](design.md) | Proposed `Store[T]` interface and envelope invariants. |
| [docs/research/bitemporal-and-go-landscape.md](research/bitemporal-and-go-landscape.md) | Primary-source facts that informed recommendations. |
| [docs/agents/issue-tracker.md](agents/issue-tracker.md) | GitHub Issues + `gh` (skills setup). |
| [docs/agents/triage-labels.md](agents/triage-labels.md) | Default triage label vocabulary. |
| [docs/agents/domain.md](agents/domain.md) | How skills consume `CONTEXT.md` and ADRs. |

Agent-facing project file: [`AGENTS.md`](../AGENTS.md).
