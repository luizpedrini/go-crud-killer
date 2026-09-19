# Docs

User-facing design for **go-crud-killer**, produced by `/setup-matt-pocock-skills` then `/grill-with-docs` (`/grilling` + `/domain-modeling`).

Start here:

1. [Grill session](grilling/session.md) — design tree, **Round 1 locked**, **Round 2 Q11–Q22 waiting**.
2. [Design](design.md) — locked Round 1 shape plus proposed `Store[T]` for Round 2.
3. [Glossary](../CONTEXT.md) — settled domain language only.
4. [ADRs](adr/) — settled, hard-to-reverse decisions.

## Layout

| Path | What it is |
| --- | --- |
| [CONTEXT.md](../CONTEXT.md) | Glossary (domain-modeling). Implementation-free. |
| [docs/adr/0001-go-library-for-entity-crud.md](adr/0001-go-library-for-entity-crud.md) | Library-only v1, module path, Go 1.23. |
| [docs/adr/0002-bitemporal-audit-on-every-entity.md](adr/0002-bitemporal-audit-on-every-entity.md) | Every Entity is bitemporal and append-only on transaction time. |
| [docs/adr/0003-terminate-keeps-history.md](adr/0003-terminate-keeps-history.md) | Terminate ends Valid time; no physical delete. |
| [docs/adr/0004-retroactive-edit-is-append-only.md](adr/0004-retroactive-edit-is-append-only.md) | Retroactive Edit splits Valid time; never rewrites Transaction time. |
| [docs/adr/0005-actor-required-on-mutations.md](adr/0005-actor-required-on-mutations.md) | Actor required on Create, Edit, Terminate. |
| [docs/adr/0006-caller-supplied-string-identity.md](adr/0006-caller-supplied-string-identity.md) | Caller-minted string Identity. |
| [docs/adr/0007-foreign-key-validation.md](adr/0007-foreign-key-validation.md) | v1 validates Foreign keys (semantics still open). |
| [docs/adr/0008-generics-over-structs.md](adr/0008-generics-over-structs.md) | `Store[T]`, no codegen. |
| [docs/adr/0009-module-owns-bitemporal-rules.md](adr/0009-module-owns-bitemporal-rules.md) | Go module owns rules; database-agnostic adapters. |
| [docs/grilling/session.md](grilling/session.md) | The interview. Open frontier lives here. |
| [docs/design.md](design.md) | Locked + proposed interface. |
| [docs/research/bitemporal-and-go-landscape.md](research/bitemporal-and-go-landscape.md) | Primary-source facts. |
| [docs/agents/issue-tracker.md](agents/issue-tracker.md) | GitHub Issues + `gh`. |
| [docs/agents/triage-labels.md](agents/triage-labels.md) | Default triage labels. |
| [docs/agents/domain.md](agents/domain.md) | How skills consume `CONTEXT.md` and ADRs. |

Agent-facing project file: [`AGENTS.md`](../AGENTS.md) (kept; not `CLAUDE.md`).
