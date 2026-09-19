# Entities are generic over caller structs; no codegen in v1

Callers define a payload struct. The module is `Store[T]`. The library owns Identity, Valid time, Transaction time, and Audit; the caller owns `T`. No schema files, no generated methods. Untyped `map`/`[]byte` payloads may exist inside an adapter codec, not as the public entity-definition story.

## Status

accepted
