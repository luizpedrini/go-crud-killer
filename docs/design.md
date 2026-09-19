# Design (Round 1 locked, Round 2 proposed)

Round 1 is locked in [`grilling/session.md`](grilling/session.md) and the ADRs. This file states those locks, then the **recommended** Round 2 shape. Round 2 recommendations are not accepted until Luiz confirms.

Vocabulary: **module**, **interface**, **seam**, **adapter**, **depth** as in the codebase-design skill. Domain words as in [`CONTEXT.md`](../CONTEXT.md).

## Locked (Round 1)

- **Library-only v1**, module `github.com/luizpedrini/go-crud-killer`, Go 1.23+. No CLI, no HTTP tool. [ADR-0001](adr/0001-go-library-for-entity-crud.md)
- Every Entity is **bitemporal**, auditable, append-only on Transaction time. [ADR-0002](adr/0002-bitemporal-audit-on-every-entity.md)
- **Terminate** ends Valid time and keeps history. No physical delete. [ADR-0003](adr/0003-terminate-keeps-history.md)
- **Edit** may be retroactive: split Valid time, never rewrite Transaction time. [ADR-0004](adr/0004-retroactive-edit-is-append-only.md)
- Omitted Valid time defaults to **`[now, unbounded)`**.
- **Actor** required on Create, Edit, Terminate. [ADR-0005](adr/0005-actor-required-on-mutations.md)
- **Identity** is a caller-supplied `string`. [ADR-0006](adr/0006-caller-supplied-string-identity.md)
- **Foreign keys** in the Payload are validated on mutation. [ADR-0007](adr/0007-foreign-key-validation.md)
- Entity definition is **`Store[T]`** over a caller struct. No codegen. [ADR-0008](adr/0008-generics-over-structs.md)
- The **Go module owns** overlap, split, Transaction-time close, and as-of selection. Adapters are **database-agnostic** and do not own those rules. [ADR-0009](adr/0009-module-owns-bitemporal-rules.md)

## Proposed deep module (Round 2 ➡️, not locked)

Callers learn a small **interface**: a payload type `T`, Identity, Actor, Create / Read / Edit / Terminate / History / List. Behind it: period splitting, Foreign key covering checks, Transaction-time close, as-of selection, optimistic conflict. The **interface is the test surface**.

```
┌──────────────────────────────────────────────┐
│  Store[T]                                    │  ← interface
│  Create Read Edit Terminate History List     │
├──────────────────────────────────────────────┤
│  Envelope + covering Foreign keys            │
│  Valid-time split / overlap                  │  ← implementation
│  Transaction-time close + append             │
├──────────────────────────────────────────────┤
│  Recording repository     Resolver (FK)      │  ← internal seams
└──────────────────────────────────────────────┘
         ▲                         ▲
   Memory adapter            Durable adapter (Postgres recommended, not locked)
```

Two adapters make the persistence **seam** real. The Resolver port is a second real seam: a registry of stores in production, a fake in tests.

## Proposed `Store[T]` (illustrative, not shipped)

```go
type Actor struct {
    ID     string
    Type   string
    Name   string
    Reason string
}

type Period struct {
    From time.Time // inclusive; zero = unbounded start
    To   time.Time // exclusive; zero = unbounded end
}

type AsOf struct {
    ValidAt       time.Time // zero → clock.Now()
    TransactionAt time.Time // zero → clock.Now() (current belief)
}

type Ref struct {
    Kind  string // e.g. "department"
    Field string // payload field name, for ErrForeignKey
    ID    string
}

type Version[T any] struct {
    ID      string
    Payload T
    Valid   Period
    Tx      Period
    Actor   Actor
}

type Clock interface {
    Now() time.Time // UTC
}

type Resolver interface {
    Covers(ctx context.Context, kind, id string, valid Period) error
}

type Store[T any] interface {
    Create(ctx context.Context, id string, payload T, valid Period, actor Actor) (Version[T], error)
    Read(ctx context.Context, id string, as AsOf) (Version[T], error)
    Edit(ctx context.Context, id string, payload T, valid Period, actor Actor) (Version[T], error)
    Terminate(ctx context.Context, id string, valid Period, actor Actor) (Version[T], error)
    History(ctx context.Context, id string) ([]Version[T], error)
    List(ctx context.Context, as AsOf) ([]Version[T], error)
}
```

Zero `valid.To` means unbounded. A convenience layer may omit `Period` and let the module fill `[clock.Now(), unbounded)` (locked Q5). There is no `Delete`. There is no Transaction-time argument on mutations.

`Refs func(T) []Ref` and optional `Validate func(T) error` are Store construction options, not methods on `T` that every payload must implement.

## Envelope invariants (locked + proposed)

Locked:

1. Identity is a non-empty string, caller-minted.
2. Actor.ID is non-empty on Create, Edit, Terminate.
3. Periods are half-open `[From, To)` and well-formed.
4. Callers cannot set Transaction time.
5. Closed Transaction-time Versions are immutable.
6. Terminate does not erase recordings.
7. Foreign keys are validated on mutation (rule for *covering* is Round 2 Q16).

Proposed (Round 2):

8. Currently believed Valid-time portions for one Identity do not overlap.
9. Create of an Identity whose current belief still covers the new Valid time is `ErrAlreadyExists`. Create after Terminate with a non-overlapping Valid-time life is allowed (Q22).
10. Edit/Terminate of a missing Identity is `ErrNotFound`.
11. Concurrent close of the same open Transaction-time Version is `ErrConflict` (Q13).
12. Foreign key covering uses **current belief** over the child's whole Valid-time period (Q16), via Resolver (Q18).

## Proposed Read coordinates

| Call | Meaning |
| --- | --- |
| `Read(id, AsOf{})` | What we **currently believe** is valid **now** |
| `Read(id, AsOf{ValidAt: march})` | Current belief about March |
| `Read(id, AsOf{ValidAt: march, TransactionAt: april})` | What we believed in April that March was |
| `History(id)` | Every Version, superseded included, transaction-time order |
| `List(AsOf{})` | Currently believed Versions valid now |

## Persistence (proposed)

- **Memory** adapter: conformance suite, no I/O.
- **Port**: append Version, close Transaction-time end, read by Identity and as-of. No `UPDATE` of historical `tx_to` except that close.
- **Postgres** adapter (Q14 ➡️ A): first durable implementation, still not the rule engine. `timestamptz`, `NULL` = unbounded. Payload codec opaque to the adapter.

## v1 is not

- A CLI, REST server, or ORM.
- Code generation.
- Physical purge, legal hold, hash-chaining.
- Multi-tenant scoping inside the library (Q21 ➡️).
- Database-native system versioning as the source of behavior.

## Next

Answer Round 2 Q11–Q22 in [`grilling/session.md`](grilling/session.md). Implementation stays blocked until the frontier is empty and you confirm shared understanding.
