# go-crud-killer

A Go library for defining simple entities and performing CRUD on them, where every entity carries full traceability, audit, and double temporal validity.

## Language

### Entity

**Entity**:
A named, identifiable thing the caller defines. Every Entity is recorded with audit, traceability, and two independent time axes.
_Avoid_: row, document, model, resource

**Identity**:
The caller-supplied `string` that distinguishes one Entity from another across Versions. The library never mints it.
_Avoid_: primary key, UUID, ULID (a possible value, not the concept)

**Payload**:
The caller-defined struct of attributes, excluding Identity and the temporal/audit envelope.
_Avoid_: data, body, model, fields

**Version**:
One recording of a Payload over a Valid-time period, believed during a Transaction-time period.
_Avoid_: revision, snapshot, row

**Actor**:
Who caused a Version to be recorded. Required on Create, Edit, and Terminate; never supplied by the library.
_Avoid_: user, system (as a default), updated_by

**Foreign key**:
A Payload field that holds another Entity's Identity, which the library validates on mutation.
_Avoid_: relation, join, association, edge

### Operations

**Create**:
The first recording of an Entity: it becomes known to the library and gains an initial Version.
_Avoid_: insert, persist, save (when those mean "write a row")

**Read**:
Retrieve an Entity (or many) under explicit time coordinates, or the defaults for "what we currently believe is valid now".
_Avoid_: get, fetch, find (as glossary terms)

**Edit**:
Record a new Version that may overlap past Valid time. Old Versions stay queryable on the transaction axis; they are never rewritten.
_Avoid_: update-in-place, mutate, patch

**Terminate**:
Record that an Entity no longer exists in the modeled world from some Valid time onward. Prior Versions remain. Not a physical delete.
_Avoid_: delete, destroy, remove, hard delete

### Time and quality

**Audit**:
Who caused a recording, and the facts needed to attribute that recording later.
_Avoid_: changelog, log line, last_modified_by (a single field is not Audit)

**Traceability**:
The ability to reconstruct what was believed about an Entity at any recording time, and what was true of it at any valid time.
_Avoid_: history (ambiguous between the two time axes), logging

**Valid time**:
When the Entity's recorded facts are true in the world the caller models. Independent of when the library learned them. May be past, present, or future. When the caller omits it, the period is `[now, unbounded)`.
_Avoid_: application time, business time, effective date (an instant, not a period), valid_from alone

**Transaction time**:
When the library recorded, superseded, or closed a particular belief about an Entity: created, edited, or terminated. Assigned by the system; not rewritten.
_Avoid_: system time, recorded_at, updated_at, created_at (single timestamps are not a period)

**Bitemporal**:
Carrying both Valid time and Transaction time as independent axes on every Version.
_Avoid_: versioned, slowly changing dimension, event sourced
