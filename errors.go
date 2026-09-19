package crudkiller

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrConflict      = errors.New("conflict")
	ErrInvalidPeriod = errors.New("invalid period")
	ErrMissingActor  = errors.New("missing actor")
	ErrForeignKey    = errors.New("foreign key")
	ErrEmptyIdentity = errors.New("empty identity")
)
