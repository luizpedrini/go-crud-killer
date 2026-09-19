# Terminate ends Valid time and keeps every recording

Terminate is not `DELETE FROM`. It records that an Entity is no longer true in the modeled world from a Valid-time instant onward, and it appends a new Transaction-time recording. Physical deletion of recordings is out of scope for the core library.

## Status

accepted

## Considered Options

- **Valid-time end, history kept (chosen).** Matches "terminated" on the transaction axis and preserves Traceability.
- **Close Transaction time only**, without a Valid-time end. That says "we no longer currently believe this exists" without saying when it stopped being true in the world.
- **Physical delete** of current and/or historical recordings. That destroys Audit.

## Consequences

There is no `Delete` method. Retention or purge, if it ever exists, is a later explicit operation, not Terminate.
