# v1 validates Foreign keys; it is not relation-free

A Payload may name another Entity's Identity. The library validates those Foreign keys on mutation. Nested payload fields that are not Identities of other Entities are just payload. Aggregates, graphs, and join APIs are still out of scope.

How existence is judged on the two time axes, how fields are declared, and how cross-kind lookup works are Round 2. This ADR only locks *that* validation belongs in v1, against the earlier recommendation to skip relations entirely.

## Status

accepted

## Considered Options

- **No relations in v1.** Smaller module. Callers reimplement "does this department exist" on every Create, and they get the time axes wrong.
- **Foreign key validation in v1 (chosen).** Keeps the CRUD interface honest for simple entities that point at other simple entities, without becoming an ORM.
