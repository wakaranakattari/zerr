// Package join groups independent errors into a single error that
// implements errors.Join semantics with zero extra allocations for
// up to four members.
//
//	type BatchError = error
//
//	var err error
//	for _, f := range files {
//		err = join.Append(err, process(f)) // nil-safe
//	}
//	if err != nil {
//		return err // errors.Is matches every member of the group
//	}
package join

import (
	"github.com/wakaranakattari/zerr/internal/multierror"
)

// Join combines errs into a single error. Nil values are discarded;
// with no non-nil values Join returns nil. A single non-nil error is
// returned unchanged, preserving its identity and attributes.
func Join(errs ...error) error {
	return multierror.Merge(errs)
}

// Append accumulates errs into err, returning a new grouped error or
// nil when the result has no non-nil members. It is nil-safe:
// Append(err, nil...) is a no-op. Use it to collect errors in a loop:
//
//	var err error
//	for _, f := range files {
//		err = join.Append(err, read(f))
//	}
//
// Append allocates once per call; collecting n errors in a loop costs
// O(n²) copied bytes overall.
func Append(err error, errs ...error) error {
	if err == nil && len(errs) == 0 {
		return nil
	}
	return multierror.Merge(append([]error{err}, errs...))
}
