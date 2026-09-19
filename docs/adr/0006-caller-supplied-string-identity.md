# Identity is a caller-supplied string

The library does not mint IDs. Callers pass a non-empty `string` Identity on Create and on every later operation. Domain IDs already exist in the systems this library is for; a built-in ULID would fight them.

## Status

accepted
