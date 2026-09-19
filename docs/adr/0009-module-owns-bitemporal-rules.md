# The Go module owns bitemporal rules; adapters are database-agnostic

Overlap, split, Transaction-time close, as-of selection, and envelope invariants live in the Go module. Persistence is a seam: adapters store and retrieve recordings and must not invent their own semantics or rewrite closed Transaction-time rows. Rules do not live in SQL:2011 system versioning, Postgres triggers, or any other datastore feature. Which SQL dialect, if any, ships in v1 is still open; the interface must not assume one.

## Status

accepted

## Considered Options

- **Database-native temporal tables own the rules.** One store gets them "for free"; every other adapter drifts; tests cannot share the implementation.
- **Go module owns the rules, adapters are dumb (chosen).** Two adapters make the seam real (memory for tests, plus at least one durable adapter). Database-agnostic means a Postgres adapter cannot be the source of truth for behavior.
