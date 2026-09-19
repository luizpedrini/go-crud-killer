# Facts that informed the grill

Primary-source notes used during `/grill-with-docs`. These are facts, not product decisions.

## Two time axes

Snodgrass / TSQL2 and the SQL/Temporal work treat **valid time** and **transaction time** as orthogonal:

- **Valid time** is when a fact is true in the modeled world. It can be past, present, or future, and the user supplies it.
- **Transaction time** is when that fact was asserted in the database. If transaction time is supported, previous database states are retained and modifications are append-only. Unlike valid time, transaction time cannot be faithfully simulated with ordinary user-writable timestamp columns, because those columns can be rewritten; the DBMS (or a library that owns the write path) must maintain them so past belief cannot be forged.

Sources:

- Snodgrass, R. T., et al. *Adding Valid Time to SQL/Temporal* / *Adding Transaction Time to SQL/Temporal* (ISO change proposals hosted at Arizona). Transaction time "identifies when data was asserted in the database" and tables with that support "grow monotonically." <https://www2.cs.arizona.edu/~rts/initiatives/sql3/mad147.pdf>
- Snodgrass / Jensen lineage: valid and transaction time are not homogeneous; a database that records both is **bitemporal**. <https://rts.cs.arizona.edu/pubs/LNCS639.pdf>
- Jensen, C. S. *Effective Timestamping in Databases*: accountability needs both "when the fact was true" and "when it was stored as current." <https://people.cs.aau.dk/~csj/Thesis/pdf/chapter40.pdf>

## SQL:2011 names

SQL:2011 standardizes the same two axes under different names:

- **Application-time period** ≈ valid time (user-managed). Supports `FOR PORTION OF` updates/deletes that split periods.
- **System-time period** ≈ transaction time (system-managed, `WITH SYSTEM VERSIONING`). Only current system-time rows can be updated or deleted; historical system-time rows are retained automatically.

A table with both periods is bitemporal. Queries can combine `FOR SYSTEM_TIME AS OF …` with an application-time predicate (`CONTAINS`, etc.).

Sources:

- Kulkarni & Michels, *Temporal features in SQL:2011*, SIGMOD Record 41(3). <https://sigmodrecord.org/publications/sigmodRecord/1209/pdfs/07.industry.kulkarni.pdf>
- PostgreSQL wiki, *SQL2011Temporal*. Application time tracks "the history of the thing out in the world"; system time tracks "when the database itself was changed." <https://wiki.postgresql.org/wiki/SQL2011Temporal>
- PostgreSQL 18/19 docs, *Temporal Tables*: temporal primary keys use a range/multirange so identity is unique *at any instant* of application time, not unique across all history. <https://www.postgresql.org/docs/current/ddl-temporal-tables.html>
- MariaDB, *Bitemporal Tables*: `PERIOD FOR application_time` plus `PERIOD FOR SYSTEM_TIME` and `WITH SYSTEM VERSIONING`. System time cannot be the target of `FOR PORTION OF`. <https://mariadb.com/docs/server/reference/sql-structure/temporal-tables/bitemporal-tables>

This repo's glossary uses **Valid time** and **Transaction time** (the academic terms that match the product brief's "when the entity is valid" / "when it was created / terminated / edited"). SQL's application time / system time are synonyms, not extra axes.

## Go landscape (what already exists)

- **GORM, ent, bun, sqlc** are snapshot CRUD (optional `UpdatedAt`). They do not give two time axes. `enthistory` and similar extensions add a history table: that is transaction time, not bitemporal CRUD.
- **lovung/gotemporal** (2020, GORM) sketched uni-temporal and bitemporal document versions. The README checkboxes for bitemporal, uni-temporal, GORM, and XORM are all still open. Not a substitute.
- **zkrebbekx/chronicle** (2026) is a Go **bitemporal change log**: opaque payloads, required actors, system-assigned transaction time, query surface on both axes, `database/sql` + Postgres ranges. It is a history module, not an entity-definition + CRUD module. Useful as a nearby existence proof, not as this library's interface.

go-crud-killer's gap, if it earns one, is the **CRUD interface for caller-defined entities** with the bitemporal envelope inside the Create / Read / Edit / Terminate path, rather than a separate log you remember to write to.

## Periods, not instants

Both axes are periods, conventionally half-open `[from, to)`. A single `created_at` / `updated_at` / `deleted_at` on a row is not double temporal validity. Create, Edit, and Terminate are events that close and open transaction-time periods (and sometimes split valid-time periods). They are not three nullable timestamp columns.
