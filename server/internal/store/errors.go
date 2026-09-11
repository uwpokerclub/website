package store

import "errors"

var ErrNotFound = errors.New("record not found")

var ErrAlreadyExists = errors.New("record already exists")

// ErrNotImplemented is returned by InMemoryStore repository methods whose logic lives
// entirely in hand-written SQL (dashboard aggregates). Reimplementing that SQL in Go
// would only test itself, not the shipped query, so those methods are not implemented
// against the in-memory store and are tested against real Postgres instead.
var ErrNotImplemented = errors.New("not implemented")
