// Package crudkiller is an in-process store for bitemporal entity CRUD.
//
// Callers define a Payload struct and a string Identity. The library owns the
// envelope — Valid time, Transaction time, Actor — and the inbound Store[T]
// port. Persistence is a Recording port; this module ships a memory adapter.
package crudkiller
