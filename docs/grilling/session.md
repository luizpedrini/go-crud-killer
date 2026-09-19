# Grill-with-docs session: go-crud-killer

`/grill-with-docs` = `/grilling` + `/domain-modeling`. This file is the interview paper trail. Glossary terms that are settled live in [`CONTEXT.md`](../../CONTEXT.md). Hard-to-reverse settled decisions live in [`docs/adr/`](../adr/).

This run was AFK: the product brief is treated as Round 0 answers. Round 1 is the current **frontier** (questions that can be asked without guessing). Recommended answers are recorded; they are **not** accepted until confirmed. Round 2 is shown only as the contingent next frontier.

The grilling session is **not finished**. The frontier is not empty. Do not implement the library until a later round closes it and you confirm shared understanding.

## Design tree

```
go-crud-killer
├── Product shape                          ← settled: Go library (ADR-0001)
│   ├── CLI / HTTP adapter later?          ← Round 1 Q1
│   └── Module path / Go version           ← Round 1 Q2
├── Domain model
│   ├── Bitemporal envelope on every Entity← settled (ADR-0002)
│   ├── Valid time vs Transaction time     ← settled (CONTEXT.md)
│   ├── Create / Read / Edit / Terminate   ← settled names; semantics Round 1 Q3–Q5
│   ├── Audit attribution                  ← Round 1 Q6
│   ├── Identity                           ← Round 1 Q7
│   └── Relations / "simple"               ← Round 1 Q8
├── Caller interface (the deep module)
│   ├── How entities are defined           ← Round 1 Q9
│   └── Operation surface                  ← blocked on Q3–Q9 → Round 2
└── Persistence
    ├── Where the bitemporal rules live    ← Round 1 Q10
    └── First production adapter           ← blocked on Q10 → Round 2
```

## Round 0 — settled from the brief

Taken as user answers, not recommendations:

1. **Name**: go-crud-killer.
2. **Kind**: new Golang **library** (the GitHub description still says "opiniated tool"; ADR-0001 follows the brief).
3. **Job**: CRUD for simple entities, built to scale, covering **entity definition** and **CRUD operations**.
4. **Minimal Entity**: full **traceability**, **audit**, and **double temporal validity**:
   - when the entity is **valid** (Valid time);
   - when it was **created / terminated / edited** (Transaction time).

Terms written to `CONTEXT.md` from this round: Entity, Create, Read, Edit, Terminate, Audit, Traceability, Valid time, Transaction time, Bitemporal.

ADRs written: [0001](../adr/0001-go-library-for-entity-crud.md), [0002](../adr/0002-bitemporal-audit-on-every-entity.md).

Facts gathered (not user decisions): [bitemporal-and-go-landscape.md](../research/bitemporal-and-go-landscape.md).

## Round 1 — current frontier

Numbered questions. One recommended answer each. Confirm, reject, or replace; then Round 2 can be asked for real.

---

❓ **Q1** - **Library-only, or library plus a tool in v1?**
The GitHub description calls this an "opiniated tool". The brief calls it a library. A CLI or HTTP admin tool can sit on the library later. Shipping both in v1 splits the interface: you design a process tool instead of a deep in-process module.

➡️ **Library only in v1.** A command or HTTP adapter is a later extra, not the product. Update the GitHub description to match.

---

❓ **Q2** - **Go version and module path?**
Remote is `github.com/luizpedrini/go-crud-killer`. Current stable Go that still gets modules right for generics + `iter`/`slog` is 1.22+; 1.23/1.24 is the practical floor if we want range-over-func iterators for History.

➡️ **Module `github.com/luizpedrini/go-crud-killer`, minimum Go 1.23.** No vanity import for v1.

---

❓ **Q3** - **What does Terminate do?**
CRUD's "Delete" usually means `DELETE FROM`. You used **terminated** on the transaction axis. Options:

- **A.** Close Valid time from a caller-supplied instant (the Entity is no longer true in the world). Transaction time records that we learned the termination now. Recordings remain.
- **B.** Close Transaction time only ("we no longer currently believe this Entity exists") without a valid-time end.
- **C.** Physical delete of current and/or historical recordings.

➡️ **A.** Terminate is a valid-time end plus a new transaction-time recording. Physical delete is out of scope for the core library (retention/purge would be a later, explicit operation if ever).

---

❓ **Q4** - **May Edit change Valid time, including the past?**
If Edit can only change "current" facts, you cannot record "we learned today that the salary in March was wrong." That is the question bitemporal Audit exists to answer. If Edit cannot be retroactive, Valid time is decoration.

➡️ **Yes. Edit may supply a Valid-time period that overlaps the past.** Overlap is resolved by splitting/closing valid-time portions on the *new* transaction-time recording, never by rewriting the old transaction-time recording. That overlap-split behavior is the first thing a prototype should prove if Q4 is accepted.

---

❓ **Q5** - **Default Valid time when the caller omits it?**
Callers of "simple CRUD" will not always think in periods.

- **A.** Required: every Create/Edit/Terminate takes an explicit Valid-time period.
- **B.** Default `[now, unbounded)` on Create/Edit, and Terminate defaults to `[now, unbounded)` as the portion that ends.
- **C.** Valid time is always unbounded ("forever") and only Transaction time is used unless the caller opts in. This contradicts ADR-0002's "every entity is bitemporal" in spirit: the axis exists but is unused.

➡️ **B.** Defaults keep simple CRUD one-line. The envelope is still always stored. Callers who care about the past pass an explicit period.

---

❓ **Q6** - **Is Audit's actor required on every mutation?**
A nullable `updated_by` is how audit trails go missing. Chronicle (a nearby Go bitemporal log) requires an actor and refuses an ambient "system" default.

➡️ **Required Actor on Create, Edit, and Terminate.** Type is a small struct (`ID` required; `Type`/`Name`/`Reason` optional). Reads do not take an Actor. No hidden default actor in the library.

---

❓ **Q7** - **Who mints Identity?**
If the library generates ULIDs, callers who already have domain IDs fight it. If the library requires caller IDs, "simple" Create needs an ID source.

➡️ **Caller supplies Identity.** The library treats it as an opaque comparable (start with `string`). No ID generator in v1.

---

❓ **Q8** - **How simple is "simple entity"? Relations?**
Temporal foreign keys (the referenced Entity must exist *for the whole valid-time period*) are a full second product. v1 can still "scale" as many independent Entities of one kind.

➡️ **v1 is one Entity kind at a time: Identity + Payload + envelope. No relations, no temporal foreign keys, no aggregates.** Nested payload fields are just payload. Cross-entity consistency is the caller's problem until a later ADR.

---

❓ **Q9** - **How do callers define an Entity?**
Options:

- **A.** Typed Go struct + a small envelope the library owns. `Store[T]` via generics. No codegen.
- **B.** Codegen (ent-style schema → methods). Heavier, more complete later.
- **C.** Untyped `map` / `[]byte` payloads (chronicle-style). Fast to persist, weak as an *entity definition* library.

➡️ **A.** Generics over a caller struct. The library owns Identity + Valid time + Transaction time + Audit; the caller owns the payload struct. Codegen is a later option if struct tags or schema files become necessary. Untyped payloads are an internal codec concern, not the public entity-definition story.

---

❓ **Q10** - **Where do the bitemporal rules live?**
If Postgres application-time / system-versioning (or a trigger extension) owns the rules, the Go module is a thin SQL wrapper and other stores cannot share behavior. If the Go module owns the rules, every store must not cheat (no `UPDATE` of closed transaction-time rows).

➡️ **The Go module owns the rules** (a deep in-process module). Persistence is a **seam** with at least two **adapters** from day one: in-memory (the test stand-in) and one SQL adapter. Adapters store and retrieve recordings; they do not invent their own overlap semantics. Postgres is the recommended first SQL adapter (ranges, exclusion constraints) but is not the interface.

---

## Round 2 — contingent (do not answer yet)

Ask these only after Round 1 is confirmed. Recommendations below assume every Round 1 ➡️ is accepted; they change if you reject any.

❓ **Q11** - **Public operation surface?**
➡️ Four mutations (`Create`, `Read`, `Edit`, `Terminate`) plus `History` (all transaction-time recordings) and `AsOf` coordinates on Read (valid-at × transaction-at). No `Upsert`, no `Patch`, no bulk API in v1. List/filter of many Identities is a later deepening.

❓ **Q12** - **Clock and time zone?**
➡️ Injected `Clock` with `Now() time.Time`. All timestamps UTC. Transaction time comes from the clock (or the database transaction timestamp if the SQL adapter can do it atomically); callers cannot set it.

❓ **Q13** - **Concurrency?**
➡️ Optimistic: Edit/Terminate fail if the current transaction-time recording has already been closed. No pessimistic locks in the in-memory adapter; SQL adapter may use a unique "open transaction-time" constraint.

❓ **Q14** - **Payload validation?**
➡️ Library checks envelope invariants (periods well-formed, actor present, identity non-empty, no overlapping *current* valid-time portions for one Identity). Payload validation is the caller's (or an optional injected `Validate(T) error`).

❓ **Q15** - **SQL dialect for the first adapter?**
➡️ PostgreSQL `timestamptz` + `tstzrange` `[)` with `NULL` as unbounded, matching the in-memory zero/`ok` convention. No GORM/ent dependency.

❓ **Q16** - **Corrections vs "new facts"?**
➡️ One Edit path. An optional `Reason` on Audit is enough to distinguish "correction" from "raise" without a second verb. If you later need intent as a first-class enum, that is a new ADR.

❓ **Q17** - **Multi-tenancy?**
➡️ None in v1. Tenant is either part of Identity or a payload field the library does not interpret.

❓ **Q18** - **Error model?**
➡️ Sentinel errors: `ErrNotFound`, `ErrAlreadyExists`, `ErrConflict`, `ErrInvalidPeriod`, `ErrMissingActor`. Wrap with `%w`. No panics on the public interface.

## Proposed glossary additions (not in CONTEXT.md yet)

Add these only after the matching Round 1 answers land:

| Term | If Q | Draft definition |
| --- | --- | --- |
| **Identity** | Q7 | Caller-supplied stable key for one Entity across Versions. _Avoid_: primary key, UUID |
| **Payload** | Q9 | Caller-defined attributes, excluding Identity and the envelope. _Avoid_: data, body |
| **Version** | Q3–Q5 | One recording of Payload over a Valid-time period, believed during a Transaction-time period. _Avoid_: revision, snapshot |
| **Actor** | Q6 | Who caused a Version to be recorded. _Avoid_: user, system (as a default) |
| **As of** | Q11 | A pair of instants (valid-at, transaction-at) that select which Version Read returns. |

## Proposed ADRs (not written until confirmed)

- Library-only v1, no CLI/HTTP product (Q1).
- Caller-supplied string Identity (Q7).
- Required Actor on mutations (Q6).
- Go module owns bitemporal rules; memory + SQL adapters (Q10).
- Retroactive Edit splits valid time; never rewrites transaction time (Q4).
- v1 has no relations (Q8).

## Status

| Item | State |
| --- | --- |
| Round 0 | Settled |
| Round 1 Q1–Q10 | **Waiting on you** |
| Round 2 Q11–Q18 | Blocked |
| Implementation | Not started; grilling says do not act until shared understanding |
