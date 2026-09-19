# go-crud-killer v1 — product vision

An in-process Go library for defining simple entities and performing Create, Read, Edit, and Terminate at scale. A minimal Entity always carries **Traceability**, **Audit**, and **double temporal validity**: **Valid time** (when the Entity is true in the modeled world) and **Transaction time** (when it was created, edited, or terminated).

Module: `github.com/luizpedrini/go-crud-killer`. Language: Go 1.23+.

## Problem

Go CRUD stacks treat an entity as the current row. A timestamp and an optional history table answer “what is true now” and sometimes “what the database contained,” but they fail the audit question: **what did we believe in April that March looked like**, after a later correction. Callers who need Valid time, Transaction time, and who recorded the change reimplement that envelope once per project, and usually get the two axes wrong.

## Product

Callers define a **Payload** struct and a string **Identity**. The library is generic over that Payload (`Store[T]`). It owns the envelope — Valid time, Transaction time, **Actor** — and the operations **Create**, **Read**, **Edit**, **Terminate**, **History**, and **List**.

There is no CLI, no HTTP service, and no ORM. People reconstruct belief through the application that embeds the module.

## Goals

1. Entity definition and CRUD in one deep module.
2. Every Entity is bitemporal, auditable, and traceable.
3. Many Identities of one kind scale through keyset pagination on History and List.
4. Hexagonal architecture: persistence is a **port**. Memory (tests) and **MySQL** (standard SQL) are **adapters** on that port.

## Non-goals

- CLI, HTTP admin, or hosted service
- Code generation, schema files, or struct-tag ORMs
- Physical deletion of recordings
- Aggregates or join APIs
- Multi-tenant filters inside the library
- Callers setting Transaction time
- Upsert, Patch, or bulk APIs
- Cascade or dangling Foreign keys (Restrict only)
- Postgres as the default SQL adapter, or a Postgres `jsonb` schema
- Legal hold, hash-chaining, crypto-shredding
- Event sourcing as the public model

## Glossary

**Entity** — A named, identifiable thing the caller defines. Always recorded with Audit, Traceability, and two time axes. _Avoid_: row, document, model, resource.

**Identity** — Caller-supplied `string` that distinguishes one Entity across Versions. The library never mints it. _Avoid_: primary key, UUID (as the concept).

**Payload** — Caller-defined struct of attributes, excluding Identity and the envelope. _Avoid_: data, body.

**Version** — One recording of a Payload over a Valid-time period, believed during a Transaction-time period. _Avoid_: revision, snapshot.

**Actor** — Who caused a Version to be recorded. Required on Create, Edit, Terminate. Optional reason. _Avoid_: system (as a default).

**Foreign key** — A Payload field that holds another Entity's Identity, validated on mutation. _Avoid_: relation, join, association.

**Covering** — A Foreign key is covering when current belief says the referenced Identity’s Valid time includes every instant of the child’s Valid time. _Avoid_: exists, snapshot FK.

**Restrict** — Parent Create, Edit, or Terminate that would leave children not Covering fails with `ErrForeignKey`. _Avoid_: cascade, dangling.

**As of** — A pair of instants (valid-at, transaction-at) that selects which Version Read or List returns. Unset means the clock’s now. _Avoid_: timestamp (one axis).

**Create / Read / Edit / Terminate** — The four CRUD verbs. Terminate is not delete.

**History** — The operation that returns every Version of one Identity, including superseded. _Avoid_: using “history” as a synonym for Traceability.

**List** — The operation that returns currently believed Versions at an As of.

**Audit** — Who caused a recording, enough to attribute it later.

**Traceability** — Reconstruct what was believed at any Transaction time and what was true at any Valid time.

**Valid time** — When the facts are true in the modeled world. A zero Period means `[now, unbounded)`.

**Transaction time** — When the library created, edited, or terminated that belief. System-assigned, not rewritten.

**Bitemporal** — Both axes on every Version.

**Resolver** — The port that answers Covering across Stores and, for Restrict, who would be uncovered.

**Port** — A hexagonal interface the core depends on (recording, clock, Resolver). _Avoid_: treating MySQL as the core.

**Adapter** — A concrete implementation of a Port (MySQL, memory, frozen clock, fake Resolver).

## Architecture

Hexagonal. The core owns overlap, Valid-time split, Transaction-time close, as-of selection, envelope invariants, Covering, and Restrict. Adapters do not own those rules. The core does not speak SQL.

| Port | Role | Adapters |
| --- | --- | --- |
| **Store[T]** | Inbound port and test surface | The application |
| **Recording** | Persist and retrieve Versions; begin/commit a unit of work | Memory (tests), **MySQL** (standard SQL) |
| **Clock** | `Now()` in UTC | Wall clock, frozen clock in tests |
| **Resolver** | Covering and Restrict reverse lookup | Registry of Stores, test fake |

Callers cannot set Transaction time. An adapter may read a datastore transaction timestamp only as the core’s clock reading, not as a public argument.

**MySQL adapter** (private to the adapter, not the port): one table per Store, named by the caller; four UTC `datetime(6)` columns (`valid_from`, `valid_to`, `tx_from`, `tx_to`) with `NULL` meaning unbounded; payload as MySQL JSON the adapter does not interpret. Documented DDL and a helper for tests; production migrations stay with the caller. Not a shared table for all kinds. Exact index names are not specified here.

No GORM, ent, or database-native system versioning as the write path. Opaque cursor bytes are not specified here.

## Operations

- **Create**, **Edit**, **Terminate** require an Actor (ID required, reason optional). They always take a Period; a zero Period means `[now, unbounded)`.
- **Read**(Identity, As of) — no Actor. Unset As of means the clock’s now on both axes (current belief, valid now).
- **History**(Identity, Page) — every Version of that Identity, including superseded, ordered by Transaction-time start, keyset-paginated.
- **List**(As of, Page) — currently believed Versions at that As of, ordered by Identity, keyset-paginated.

Page: limit plus opaque cursor. Default limit 100, maximum 1000. Empty cursor is the first page. Results are Versions plus a next cursor; an empty next cursor is the last page.

There is no Delete, Upsert, Patch, or bulk API. One Edit path covers both a new fact and a correction.

Sentinel errors, wrapped: `ErrNotFound`, `ErrAlreadyExists`, `ErrConflict`, `ErrInvalidPeriod`, `ErrMissingActor`, `ErrForeignKey`. No panics on the public interface.

Envelope validation is the library’s. Payload validation is an optional function at Store construction. Foreign keys are declared with `func(T) []Ref` at construction. Tenant, if any, is part of Identity or Payload; the library does not filter by tenant.

## Invariants

1. Identity is a non-empty caller-minted string.
2. Actor ID is non-empty on Create, Edit, and Terminate. The library never invents an Actor.
3. Periods are half-open `[From, To)` and well-formed. Zero Period on a mutation is `[clock.Now(), unbounded)`.
4. Every Version has independent Valid time and Transaction time. Callers never set Transaction time. Timestamps are UTC.
5. Closed Transaction-time Versions are immutable. Edit and Terminate append; they do not rewrite.
6. Currently believed Valid-time portions for one Identity do not overlap.
7. Create of an Identity whose current belief still covers the new Valid time is `ErrAlreadyExists`. After Terminate, Create of the same Identity is allowed when Valid-time lives do not overlap.
8. Edit and Terminate of a missing Identity are `ErrNotFound`.
9. Concurrent close of the same open Transaction-time Version is `ErrConflict`.
10. A Foreign key is valid only when it is Covering under current belief over the child’s whole Valid-time period. Same-kind and cross-kind checks go through the Resolver.
11. A parent mutation that would leave children not Covering **Restricts** with `ErrForeignKey`. The Resolver asks each Store that declared Foreign keys to that parent kind who would be uncovered. Reverse check and parent write share one recording-port unit of work.
12. History and List return Versions, never IDs only.

## Expected behavior

```gherkin
Feature: Bitemporal entity CRUD
  As a Go application developer
  I embed go-crud-killer
  So that simple entities stay auditable on Valid time and Transaction time

  Background:
    Given an injected Clock frozen at 2024-04-15T12:00:00Z
    And a Department store and an Employee store on the same Resolver
    And Foreign keys on Employee are declared at Store construction as kind "department"
    And department "ops" was Created by actor "hr" with Period 2018-01-01T00:00:00Z to unbounded

  Scenario: Create records a Version with both time axes
    When actor "hr" Creates employee "alice" with payload salary 50000 and a zero Period
    Then the Version Identity is "alice"
    And Valid time is [2024-04-15T12:00:00Z, unbounded)
    And Transaction time starts at 2024-04-15T12:00:00Z and is open
    And the Actor is "hr"
    And timestamps are UTC

  Scenario: Default Period is now to unbounded
    When actor "hr" Creates employee "bob" with a zero Period
    Then Valid time From equals the Clock's Now
    And Valid time To is unbounded

  Scenario: Explicit Valid time may lie in the past
    When actor "hr" Creates employee "cara" with Period 2024-01-01T00:00:00Z to unbounded
    Then Valid time is [2024-01-01T00:00:00Z, unbounded)
    And Transaction time still starts at 2024-04-15T12:00:00Z

  Scenario: Read current belief valid now
    Given employee "alice" exists with salary 50000 valid from 2024-04-15T12:00:00Z unbounded
    When I Read "alice" with an unset As of
    Then the Payload salary is 50000

  Scenario: Read current belief at a Valid-time instant
    Given employee "alice" is valid [2024-03-01T00:00:00Z, unbounded) with salary 50000
    When I Read "alice" As of valid-at 2024-03-15T00:00:00Z and unset transaction-at
    Then the Payload salary is 50000

  Scenario: Read what we believed in April that March was
    Given on 2024-03-01T00:00:00Z actor "hr" Created "alice" with salary 50000 valid [2024-03-01T00:00:00Z, unbounded)
    And the Clock is later 2024-04-15T12:00:00Z
    And actor "hr" Edits "alice" with salary 55000 valid [2024-03-01T00:00:00Z, unbounded)
    When I Read "alice" As of valid-at 2024-03-15T00:00:00Z and transaction-at 2024-04-01T00:00:00Z
    Then the Payload salary is 50000
    When I Read "alice" As of valid-at 2024-03-15T00:00:00Z and transaction-at 2024-04-16T00:00:00Z
    Then the Payload salary is 55000

  Scenario: Retroactive Edit splits Valid time and never rewrites Transaction time
    Given "alice" has an open Version with salary 50000 valid [2024-01-01T00:00:00Z, unbounded) recorded at 2024-01-15T00:00:00Z
    And the Clock is 2024-04-15T12:00:00Z
    When actor "hr" Edits "alice" with salary 55000 valid [2024-03-01T00:00:00Z, unbounded)
    Then History of "alice" still contains the Version recorded at 2024-01-15T00:00:00Z with salary 50000
    And that Version's Transaction time is closed and not rewritten in place
    And a new Version with salary 55000 has Transaction time starting at 2024-04-15T12:00:00Z
    And currently believed Valid-time portions for "alice" do not overlap

  Scenario: Terminate ends Valid time and keeps History
    Given employee "alice" is currently believed valid
    When actor "hr" Terminates "alice" with Period 2024-04-15T12:00:00Z to unbounded
    Then Read of "alice" with unset As of is ErrNotFound
    And History of "alice" still returns every prior Version
    And no recording is physically deleted

  Scenario: Restrict parent Terminate while children require Covering
    Given employee "alice" has Foreign key department "ops" and is currently believed valid [2024-04-15T12:00:00Z, unbounded)
    When actor "hr" Terminates department "ops" with a zero Period
    Then the result is ErrForeignKey
    And department "ops" remains currently believed valid
    When actor "hr" Terminates employee "alice" with a zero Period
    And actor "hr" Terminates department "ops" with a zero Period
    Then department "ops" is no longer currently believed valid

  Scenario: Restrict and the parent write share one unit of work
    Given employee "alice" currently requires Covering of department "ops"
    When actor "hr" Terminates department "ops" while a concurrent Create would attach another employee to "ops"
    Then either the Terminate is ErrForeignKey or the concurrent Create is ErrForeignKey or ErrConflict
    And no child is left not Covering under current belief

  Scenario: Create after Terminate of the same Identity
    Given actor "hr" Terminated "alice" so Valid time ends at 2024-06-01T00:00:00Z
    When actor "hr" Creates "alice" with Period 2024-07-01T00:00:00Z to unbounded
    Then Create succeeds
    When actor "hr" Creates "alice" with Period 2024-05-01T00:00:00Z to unbounded
    Then Create is ErrAlreadyExists

  Scenario: Foreign key must Cover the child's whole Valid time
    Given department "ops" is currently believed valid only [2018-01-01T00:00:00Z, 2019-01-01T00:00:00Z)
    When actor "hr" Creates employee "alice" with Foreign key "ops" and Period 2020-01-01T00:00:00Z to unbounded
    Then Create is ErrForeignKey

  Scenario: Covering uses current belief
    Given department "ops" is currently believed valid [2018-01-01T00:00:00Z, unbounded)
    When actor "hr" Creates employee "alice" with Foreign key "ops" and Period 2020-01-01T00:00:00Z to unbounded
    Then Create succeeds

  Scenario: History is keyset-paginated Versions ordered by Transaction-time start
    Given "alice" has three Versions recorded at T1, T2, and T3
    When I History "alice" with Page limit 2 and an empty cursor
    Then I receive two Versions in Transaction-time start order
    And the next cursor is not empty
    When I History "alice" with Page limit 2 and that cursor
    Then I receive the remaining Version
    And the next cursor is empty

  Scenario: List is keyset-paginated Versions ordered by Identity
    Given currently believed employees "alice", "bob", and "cara"
    When I List with unset As of, Page limit 2, and an empty cursor
    Then I receive Versions for "alice" and "bob" in Identity order
    And the next cursor is not empty
    When I List with that cursor
    Then I receive the Version for "cara"
    And the next cursor is empty

  Scenario: Actor is required on every mutation
    When someone Creates, Edits, or Terminates without an Actor ID
    Then the result is ErrMissingActor
    And the library does not supply a default Actor

  Scenario Outline: Create with a zero Period uses now to unbounded
    Given the Clock's Now is 2024-04-15T12:00:00Z
    When actor "hr" Creates employee "<id>" with a zero Period
    Then Valid time is [2024-04-15T12:00:00Z, unbounded)

    Examples:
      | id    |
      | alice |
      | bob   |

  Scenario: Edit with a zero Period uses now to unbounded
    Given employee "alice" is currently believed valid
    And the Clock's Now is 2024-04-15T12:00:00Z
    When actor "hr" Edits "alice" with a zero Period
    Then the new Version's Valid time uses From 2024-04-15T12:00:00Z and To unbounded

  Scenario: Terminate with a zero Period ends Valid time from now
    Given employee "alice" is currently believed valid
    And the Clock's Now is 2024-04-15T12:00:00Z
    When actor "hr" Terminates "alice" with a zero Period
    Then Valid time of current belief ends at 2024-04-15T12:00:00Z

  Scenario: Concurrent Edits conflict
    Given "alice" has one open Transaction-time Version
    When two Actors Edit "alice" against that same open Version at the same time
    Then one Edit succeeds
    And the other is ErrConflict
    And the succeeded Version remains the only open Transaction-time Version

  Scenario: Clock is injected and Transaction time is not caller-set
    Given a Clock frozen at 2024-04-15T12:00:00Z
    When actor "hr" Creates employee "alice" with any Period
    Then every timestamp on the Version is UTC
    And Transaction time start equals the Clock's Now
    And there is no mutation argument that sets Transaction time

  Scenario: The same behavior holds on memory and MySQL adapters
    Given the recording port is backed by the memory adapter or the MySQL adapter
    Then every scenario in this Feature holds through Store[T]
```

## Conformance

Behavior is defined at `Store[T]`. The same suite runs on the memory adapter and the MySQL adapter. Tests inject a Clock and may fake the Resolver. They do not assert on SQL tables through `Store[T]`.
