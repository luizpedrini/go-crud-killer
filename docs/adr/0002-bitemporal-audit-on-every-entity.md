# Every entity is bitemporal, auditable, and append-only on the transaction axis

A minimal Entity is not a current-row CRUD record. It always carries Valid time (when the facts are true in the modeled world) and Transaction time (when the library created, edited, or terminated that belief), plus enough Audit to attribute the recording. Transaction-time history is not overwritten.

## Status

accepted

## Considered Options

- **Snapshot CRUD plus optional `updated_at`.** The obvious Go CRUD library. It cannot answer "what did we believe in April that March looked like" after a later Edit, so it fails the traceability requirement.
- **Valid time only (SCD2 / application-time).** Answers "when was this true in the world" but lets a correction rewrite the only timeline, so Audit cannot reconstruct prior belief.
- **Transaction time only (system-versioned / audit table).** Answers "what did the database contain" but cannot record that a fact was true last year and learned today.
- **Bitemporal envelope on every Entity (chosen).** Two independent periods. Valid time is a caller concern; Transaction time is system-assigned and closed, never rewritten. Physical erasure of recordings is out of scope for the core model.

## Consequences

Typical "update the row" and "delete the row" mental models are wrong here. Edit and Terminate record new belief; they do not destroy old recordings. Storage adapters must preserve that invariant.
