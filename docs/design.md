# Proposed design (pending Round 1)

This is the recommended shape of go-crud-killer **if every Round 1 answer in [`grilling/session.md`](grilling/session.md) is accepted**. It is not an accepted ADR. The grilling session is still open; this file exists so the recommendation is reviewable in one place.

Vocabulary: **module**, **interface**, **seam**, **adapter**, **depth** as in the codebase-design skill. Domain words as in [`CONTEXT.md`](../CONTEXT.md).

## The deep module

Callers should learn a small **interface**: define a payload type, supply Identity and Audit, call Create / Read / Edit / Terminate. Behind that interface lives period splitting, overlap, transaction-time closing, as-of selection, and concurrency. If deleting the module only removes pass-through SQL, it failed. If callers would each reimplement bitemporal CRUD, it earned its keep.

```
┌──────────────────────────────────────────┐
│  Store[T]  Create Read Edit Terminate    │  ← interface (and test surface)
│            History, AsOf coordinates     │
├──────────────────────────────────────────┤
│  Envelope invariants                     │
│  Valid-time split / overlap              │  ← implementation (in-process)
│  Transaction-time close + append         │
│  Audit attachment                        │
├──────────────────────────────────────────┤
│  Recording repository  (internal seam)   │
└──────────────────────────────────────────┘
         ▲                    ▲
   Memory adapter       SQL adapter (Postgres first)
```

Two adapters make the persistence **seam** real (test + production). The rules do not live in Postgres system versioning, because then the memory adapter would be a second, drifting implementation.

## Entity definition

The caller owns the payload struct. The library owns the envelope.

```go
// Illustrative, not shipped.

type Actor struct {
    ID     string
    Type   string // optional, e.g. "user" | "service"
    Name   string
    Reason string
}

type Period struct {
    From time.Time // inclusive; zero means unbounded start (rare)
    To   time.Time // exclusive; zero means unbounded end
}

type AsOf struct {
    ValidAt       time.Time // zero → clock.Now()
    TransactionAt time.Time // zero → clock.Now() (current belief)
}

type Version[T any] struct {
    ID         string
    Payload    T
    Valid      Period
    Tx         Period
    Actor      Actor
    RecordedOp string // "create" | "edit" | "terminate"
}

type Store[T any] interface {
    Create(ctx context.Context, id string, payload T, valid Period, actor Actor) (Version[T], error)
    Read(ctx context.Context, id string, as AsOf) (Version[T], error)
    Edit(ctx context.Context, id string, payload T, valid Period, actor Actor) (Version[T], error)
    Terminate(ctx context.Context, id string, valid Period, actor Actor) (Version[T], error)
    History(ctx context.Context, id string) ([]Version[T], error)
}
```

`Period` zero `To` means unbounded. Omitting `valid` at the convenience wrapper layer defaults to `[now, unbounded)` (Round 1 Q5). `Terminate`'s `valid` is the valid-time interval that ends (default `[now, unbounded)`), not a physical delete.

There is no `Delete` method. There is no way to set Transaction time.

## Envelope invariants (library-enforced)

1. Identity is non-empty.
2. Actor.ID is non-empty on Create, Edit, Terminate.
3. Periods are well-formed: `From` before `To` when both are bounded; half-open `[From, To)`.
4. Transaction time is assigned by the module (clock or datastore). Callers cannot pass it in.
5. For a given Identity, currently believed valid-time portions (open transaction-time end) do not overlap.
6. Edit/Terminate of a missing Identity is `ErrNotFound`. Create of an Identity that still has an open current recording is `ErrAlreadyExists` (re-Create after Terminate is a later Round 2 question; recommendation: allowed, it is a new valid-time life, not an overwrite).
7. Closed transaction-time recordings are immutable.

## Read coordinates

| Call | Meaning |
| --- | --- |
| `Read(id, AsOf{})` | What we **currently believe** is valid **now** |
| `Read(id, AsOf{ValidAt: march})` | Current belief about March |
| `Read(id, AsOf{ValidAt: march, TransactionAt: april})` | What we believed in April that March was |
| `History(id)` | Every recording, including superseded, in transaction-time order |

That last as-of pair is the Audit/traceability test. Snapshot CRUD cannot pass it.

## Persistence adapters

- **Memory**: slice/map of recordings. The conformance suite runs here. No I/O.
- **Postgres**: one table per store (or one table with a kind column if multiple `T` share a database). `valid_from`/`valid_to`, `tx_from`/`tx_to` as `timestamptz`, `NULL` = unbounded; generated `tstzrange` for exclusion on *open* tx rows. Payload as `jsonb` or a codec the adapter does not interpret.

Adapters insert and close rows the module has already decided. They do not accept arbitrary `UPDATE` of historical `tx_to`.

## What v1 explicitly is not

- A CLI, REST server, or ORM.
- Relational/temporal foreign keys.
- Event sourcing as the public model (an adapter *may* append events internally; callers still see Versions).
- Physical purge, legal hold, hash-chaining, or crypto-shredding.
- Multi-tenant scoping.
- Code generation.

## Scaling

"Scalable" here means: the interface stays small while recordings grow. History is keyset-paginated before it is offset-paginated (Round 2 detail). The SQL adapter uses range indexes and will not table-scan on `Read`/`History` for a single Identity. Horizontal scale is "many Identities, one kind", not "a graph of kinds".

## Next

Answer Round 1 in [`grilling/session.md`](grilling/session.md). Accepted answers move into `CONTEXT.md` / `docs/adr/`. Rejected recommendations rewrite this file. Implementation starts only after you confirm the frontier is done.
