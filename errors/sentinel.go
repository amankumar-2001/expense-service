package errors

import "errors"

// ErrNotFound is returned by repositories when a row does not exist. Services map
// it to the appropriate HTTP error (usually a 404).
var ErrNotFound = errors.New("not found")

// Is reports whether err is or wraps target.
func Is(err, target error) bool { return errors.Is(err, target) }
