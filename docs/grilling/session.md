# Grill-with-docs session: go-crud-killer

`/grill-with-docs` = `/grilling` + `/domain-modeling`. This file is the interview paper trail. Glossary terms that are settled live in [`CONTEXT.md`](../../CONTEXT.md). Hard-to-reverse settled decisions live in [`docs/adr/`](../adr/).

Luiz locked Round 1 on 2026-09-19. Round 2 is the current **frontier**. Recommended answers below are **not** accepted until confirmed. Do not implement the library until the frontier is empty and you confirm shared understanding.

## Design tree

```
go-crud-killer
├── Product shape                          ✓ library-only v1 (ADR-0001)
│   ├── CLI / HTTP in v1                   ✓ no
│   └── Module path / Go version           ✓ github.com/luizpedrini/go-crud-killer, Go 1.23
├── Domain model
│   ├── Bitemporal envelope                ✓ ADR-0002
│   ├── Terminate                          ✓ valid-time end, keep history (ADR-0003)
│   ├── Edit                               ✓ retroactive, append-only (ADR-0004)
│   ├── Default Valid time                 ✓ [now, unbounded)
│   ├── Actor                              ✓ required on mutations (ADR-0005)
│   ├── Identity                           ✓ caller string (ADR-0006)
│   └── Foreign key                        ✓ validate on mutation (ADR-0007); semantics → Round 2
├── Caller interface
│   ├── Entity definition                  ✓ Store[T], no codegen (ADR-0008)
│   ├── Operation surface                  ← Round 2
│   ├── Clock                              ← Round 2
│   ├── Concurrency                        ← Round 2
│   ├── Errors                             ← Round 2
│   └── Foreign key mechanics              ← Round 2
└── Persistence
    ├── Who owns the rules                 ✓ Go module, DB-agnostic adapters (ADR-0009)
    └── Which adapters ship                ← Round 2 (Postgres is not locked)
```

## Round 0 — locked (product brief)

1. **Name**: go-crud-killer.
2. **Kind**: Golang **library**.
3. **Job**: CRUD for simple entities, covering entity definition and CRUD operations.
4. **Minimal Entity**: Traceability, Audit, and double temporal validity (Valid time + Transaction time).

## Round 1 — locked 2026-09-19 (Luiz)

Each item is the locked decision. The original ➡️ was the recommendation; **Q8 and the Postgres part of Q10 diverged**.

| Q | Locked |
| --- | --- |
| **Q1** | Library-only v1. No CLI, no HTTP tool. [ADR-0001](../adr/0001-go-library-for-entity-crud.md) |
| **Q2** | Module `github.com/luizpedrini/go-crud-killer`, minimum Go 1.23. [ADR-0001](../adr/0001-go-library-for-entity-crud.md) |
| **Q3** | Terminate = end Valid time, keep history. No physical delete. [ADR-0003](../adr/0003-terminate-keeps-history.md) |
| **Q4** | Edit may be retroactive: split Valid time, never rewrite Transaction time. [ADR-0004](../adr/0004-retroactive-edit-is-append-only.md) |
| **Q5** | Default Valid time when omitted: `[now, unbounded)`. |
| **Q6** | Actor required on every mutation. [ADR-0005](../adr/0005-actor-required-on-mutations.md) |
| **Q7** | Identity minted by the caller as a `string`. [ADR-0006](../adr/0006-caller-supplied-string-identity.md) |
| **Q8** | **Foreign key validation** in v1 (not “no relations”). [ADR-0007](../adr/0007-foreign-key-validation.md). How existence is judged is Round 2. |
| **Q9** | Generics over structs. No codegen. [ADR-0008](../adr/0008-generics-over-structs.md) |
| **Q10** | Bitemporal rules owned by the Go module. **Database-agnostic adapters**, not DB-owned rules. [ADR-0009](../adr/0009-module-owns-bitemporal-rules.md). Postgres-as-first-SQL-adapter was *not* locked. |

Terms added to `CONTEXT.md` this round: Identity, Payload, Version, Actor, Foreign key; Edit/Terminate/Valid time sharpened.

## Round 2 — current frontier

Prerequisites are settled. Ask all of these now. Questions that hang off these answers (referential action on Terminate, History pagination, SQL schema) wait for Round 3.

---

❓ **Q11** - **Public operation surface?**
The locked verbs are Create, Read, Edit, Terminate. Traceability also needs a way to see superseded Versions and to pin both clocks.

- **A.** `Create`, `Read(id, AsOf)`, `Edit`, `Terminate`, plus `History(id)` (every Version, including superseded).
- **B.** A plus `List(AsOf)` of Identities currently believed valid at that as-of (CRUD without List is an incomplete simple store).
- **C.** A plus bulk Create/Edit and `Upsert` / `Patch`.

➡️ **B.** Five single-Identity operations plus `List`. No `Upsert`, `Patch`, or bulk in v1. `AsOf` is a pair `(ValidAt, TransactionAt)`; zero means the clock's now. `History` is per Identity, not a global log. Pagination of `List`/`History` waits until this lands.

---

❓ **Q12** - **Clock and time zone?**
Default Valid time is `[now, unbounded)`. Transaction time is system-assigned (ADR-0002, ADR-0009). Tests cannot freeze `time.Now`. SQL adapters sometimes prefer the datastore transaction timestamp so every row in one commit shares one Transaction-time start.

- **A.** Injected `Clock` (`Now() time.Time`), all timestamps UTC, callers cannot set Transaction time. An adapter may source `Now` from its transaction as long as the *module* still assigns the period.
- **B.** Always `time.Now().UTC()` inside the module. Simpler; tests get flaky or need to sleep.
- **C.** Callers pass Transaction time. Destroys Audit (they can backdate belief).

➡️ **A.** Injected `Clock`, UTC, Transaction time not on the public mutation arguments. Adapter-provided transaction time is an adapter detail, not a second clock callers see.

---

❓ **Q13** - **Concurrency?**
Two Edits of the same Identity can race: both read an open Transaction-time Version, both try to close it. Foreign key checks can race with Terminate of the referenced Entity.

- **A.** Optimistic: mutation fails if the open Transaction-time Version it based itself on is already closed (`ErrConflict`). No pessimistic locks in the memory adapter. Durable adapters may enforce “at most one open Transaction-time row per Identity” as a constraint, still reporting `ErrConflict`.
- **B.** Pessimistic lock per Identity on every mutation.
- **C.** Last write wins: close whatever is open and append. Silently drops a concurrent Actor's belief.

➡️ **A.** Optimistic `ErrConflict`. Last-write-wins would hide a lost Edit. Pessimistic locks are an adapter optimization later, not the interface.

---

❓ **Q14** - **Which adapters ship in v1? Postgres?**
Rules are database-agnostic (ADR-0009). Two adapters make the persistence seam real. Postgres has ranges and exclusion constraints; it must still not own the rules.

- **A.** In-memory adapter (conformance suite) + a dialect-free `RecordingStore` port + a Postgres adapter as the first durable implementation (`timestamptz`, `NULL` = unbounded). No GORM/ent. Other SQL dialects are later adapters behind the same port.
- **B.** In-memory only in v1. Durable adapters after the interface is proven.
- **C.** Postgres-only, no memory adapter. Breaks the seam and the test story.

➡️ **A.** Memory + Postgres, same port. Postgres is the first *durable adapter*, not a second rule engine. If this is too much for v1, take **B** — do not take **C**.

---

❓ **Q15** - **Error model?**
Callers need to distinguish “not there”, “already there”, “someone else just edited”, “bad period”, “no Actor”, and “Foreign key failed” without string-matching.

➡️ **Sentinel errors**, wrapped with `%w`: `ErrNotFound`, `ErrAlreadyExists`, `ErrConflict`, `ErrInvalidPeriod`, `ErrMissingActor`, `ErrForeignKey`. No panics on the public interface. `ErrForeignKey` names the field and the missing Identity; it does not invent a second error type per kind.

---

❓ **Q16** - **When does a Foreign key count as valid?**
Locked: the library validates Foreign keys on mutation. Not locked: *existence on which axis*.

Scenario: Create Employee Alice, Valid `[2020-01-01, unbounded)`, Foreign key `DepartmentID = "ops"`.

- Department `ops` currently believed valid only `[2018-01-01, 2019-01-01)` → should Create fail?
- Department `ops` valid `[2018-01-01, unbounded)` as of today, but a Read as of 2019-06-01 at Transaction time 2019-06-01 showed no `ops` (it was Created later, backdated)? That is, current belief about 2020 vs belief-at-the-time?

SQL:2011 temporal foreign keys require the referenced key to exist **throughout the referencing row's application-time period**, in the current system-time state.

- **A.** **Covering, current belief.** For the child's Valid-time period (after defaulting), every instant must be covered by a currently believed Version of the referenced Identity. Retroactive Create/Edit uses *today's* belief about that past, not the belief that existed then. Matches SQL:2011 application-time FK.
- **B.** **Instant, current belief.** Only the start instant (or `ValidAt = now`) of the child must resolve. Cheaper; allows Alice `[2020, ∞)` to point at a department that ended in 2021.
- **C.** **Snapshot.** The referenced Identity has any open Version right now, ignoring Valid time. Fights ADR-0002: a bitemporal child with a snapshot parent.

➡️ **A.** Covering on current belief. Otherwise “simple” Foreign keys lie about Valid time. Belief-at-the-time checks are a later Read concern, not a mutation guard.

---

❓ **Q17** - **How is a Foreign key declared on `T`?**
No codegen (ADR-0008). The module must know which Payload fields are Identities of which Entity **kind**.

- **A.** A function the caller passes when constructing `Store[T]`: `func(T) []Ref` where `Ref` is `{Kind, Field, ID string}`. Explicit, testable, no reflection required on the happy path.
- **B.** Struct tags (`crud:"fk,kind=department"`). Looks like an ORM, needs reflection, fails late.
- **C.** Same-kind only, field named `ParentID`. Too small given Q8.

➡️ **A.** Caller-supplied `Refs(T) []Ref` (or an interface `T` may optionally implement). Tags are a later convenience, not v1.

---

❓ **Q18** - **Same-kind only, or cross-kind Foreign keys?**
`Store[T]` sees one payload type. Employee → Manager is same kind. Employee → Department is two stores.

- **A.** Cross-kind via an injected **Resolver** port: `Covers(ctx, kind, id, valid Period) error`, implemented by a registry of stores (or a test fake). Same-kind FKs use the same Store through that port too, so there is one check.
- **B.** Same-kind only in v1. Cross-kind is the caller's problem. Shrinks Q8 a lot.
- **C.** One mega-store of untyped payloads. Contradicts ADR-0008.

➡️ **A.** Resolver port. Without it, Foreign key validation cannot see Department from `Store[Employee]`. The port is justified by two adapters: the real registry and an in-memory fake for tests.

---

❓ **Q19** - **Payload validation beyond the envelope?**
Envelope invariants are the library's (non-empty Identity, Actor, well-formed periods, no overlapping *currently believed* Valid-time portions, Foreign keys per Q16). `T` may have its own rules (email format, money ≥ 0).

➡️ **Optional `Validate(T) error` injected at Store construction.** Nil means payload is opaque. The library does not tag-parse `T`.

---

❓ **Q20** - **Correction vs “new fact”?**
Retroactive Edit (ADR-0004) covers both “Alice's March salary was wrong” and “Alice's salary rose in June”.

➡️ **One Edit path.** Optional `Reason` on Actor is enough in v1. A first-class intent enum is a later ADR if Audit queries need it.

---

❓ **Q21** - **Multi-tenancy?**
Nothing in Round 1 named tenants.

➡️ **None in v1.** A tenant is either part of Identity (`"acme:alice"`) or a Payload field the library does not interpret. No hidden `WHERE tenant_id` in adapters.

---

❓ **Q22** - **Create again after Terminate?**
Scenario: Terminate Alice with Valid end 2021-01-01. In 2022 someone Creates Identity `"alice"` again with Valid `[2022-01-01, unbounded)`.

- **A.** Allowed: a new Valid-time life for the same Identity. History contains both lives. Create fails only if a currently believed Version still covers overlapping Valid time (`ErrAlreadyExists`).
- **B.** Forbidden forever: Identity is one life. Rehire needs a new Identity.
- **C.** Create after Terminate is an Edit. Hides the verb the Actor chose.

➡️ **A.** Same Identity, non-overlapping current Valid-time lives. Forbidding it forever makes Terminate a dead end for domain IDs that naturally return (employees, accounts).

## Round 3 — blocked (do not answer yet)

These wait on Round 2:

- Referential action when Terminate (or retroactive Edit) of a referenced Entity would leave children uncovered: restrict vs cascade-Terminate vs allow dangling (needs Q16, Q18).
- `List`/`History` pagination and order (needs Q11).
- Postgres column types / `tstzrange` vs portable columns (needs Q14 = A).
- Whether convenience wrappers omit `Period` in the Go signatures or only default internally (Q5 is locked; signature shape follows Q11–Q12).

## Status

| Item | State |
| --- | --- |
| Round 0 | Locked |
| Round 1 Q1–Q10 | **Locked** (Luiz, 2026-09-19) |
| Round 2 Q11–Q22 | **Waiting on you** |
| Round 3 | Blocked |
| Implementation | Not started |
