package store

import "errors"

var ErrNotFound = errors.New("record not found")

// ErrTransactionConflict means the in-memory transaction snapshot became stale
// before it could be committed.
var ErrTransactionConflict = errors.New("transaction conflict")
