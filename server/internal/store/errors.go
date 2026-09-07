package store

import "errors"

var ErrNotFound = errors.New("record not found")
var ErrConflict = errors.New("conflict")

// ErrTransactionConflict means the in-memory transaction snapshot became stale
// before it could be committed.
var ErrTransactionConflict = errors.New("transaction conflict")
