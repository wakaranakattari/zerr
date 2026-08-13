// Package must provides panic-based error flow for boundaries.
//
// Traditional Go threading of error values through every layer of a
// call stack is the one genuinely painful part of error handling; must
// lets an inner layer panic with the full error value, and Catch at a
// boundary turns it back into an explicit error. The panic carries the
// original error with its chain, attributes and (in herr_trace builds)
// stack intact.
//
//	func handle(r *http.Request) (err error) {
//		defer must.Catch(&err)
//		user := must.Must(fetchUser(r))
//		_ = user
//		return nil
//	}
//
// Use it sparingly: on request boundaries, batch jobs, and anywhere
// the middle of the call stack should not be cluttered with plumbing.
package must

import "github.com/wakaranakattari/zerr"

// Must returns v when err is nil and panics with err otherwise. It is
// for use at boundaries where an error can only mean a programming
// mistake: the panic carries the original error value, and Catch
// restores it with its full chain and stack intact.
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// MustErr panics with err if it is non-nil. Call it in defer-free
// positions when only the error needs to be asserted.
func MustErr(err error) {
	if err != nil {
		panic(err)
	}
}

// Catch recovers a panic raised by Must or MustErr (or any other
// panic) and stores the result in *errp, turning panic-based flow
// control back into explicit errors at the boundary:
//
//	func loadAll() (err error) {
//		defer must.Catch(&err)
//		return loadAllInner()
//	}
//
// A panic value that is an error is stored as-is, preserving its type,
// attributes and stack; any other panic value is wrapped in a zerr
// node via Newf. Catch re-panics when errp is nil and does nothing
// when no panic occurred.
func Catch(errp *error) {
	r := recover()
	if r == nil {
		return
	}
	if errp == nil {
		panic(r)
	}
	if er, ok := r.(error); ok {
		*errp = er
		return
	}
	*errp = zerr.Newf("panic: %v", r)
}
