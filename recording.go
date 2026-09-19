package crudkiller

import "context"

// Recording persists and retrieves Versions and begins a unit of work.
// Memory and MySQL are adapters on this port. The core owns overlap and
// envelope rules; adapters do not.
type Recording[T any] interface {
	Begin(ctx context.Context) (UnitOfWork[T], error)
}

// UnitOfWork is one recording-port transaction. Commit persists; Rollback
// discards. After either, the unit of work is finished.
type UnitOfWork[T any] interface {
	// Versions returns every recording of identity, including superseded.
	Versions(ctx context.Context, identity string) ([]Version[T], error)
	Insert(ctx context.Context, v Version[T]) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
