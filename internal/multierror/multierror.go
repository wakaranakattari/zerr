// Package multierror implements the single anchored error node that
// groups several independent errors into one, mirroring errors.Join
// while keeping up to four children in its own allocation.
//
// The package is internal: the public surface is the zerr/join
// subpackage. The zerr core imports it for traversal and rendering.
package multierror

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
)

// Error groups several independent errors into one node. It implements
// the error interface plus Unwrap() []error, so errors.Is and
// errors.AsType traverse every member of the group.
type Error struct {
	buf   [4]error
	n     int
	extra []error
}

// Merge combines errs into a single group. Nil values are discarded;
// with no non-nil values it returns nil. A single non-nil error is
// returned unchanged, preserving its identity and attributes. Up to
// four non-nil children fit in the node's own allocation.
func Merge(errs []error) error {
	hasNil := false
	n := 0
	var first error
	for _, e := range errs {
		if e == nil {
			hasNil = true
			continue
		}
		n++
		if first == nil {
			first = e
		}
	}
	switch {
	case n == 0:
		return nil
	case n == 1:
		if !hasNil {
			return first
		}
		for _, e := range errs {
			if e != nil {
				return e
			}
		}
	case n <= 4 && !hasNil:
		j := &Error{}
		copy(j.buf[:n], errs)
		j.n = n
		return j
	}
	joined := make([]error, 0, n)
	for _, e := range errs {
		if e != nil {
			joined = append(joined, e)
		}
	}
	return &Error{extra: joined}
}

// Error joins the children's messages with "; " on a single line.
func (e *Error) Error() string {
	errs := e.Unwrap()
	parts := make([]string, 0, len(errs))
	for _, er := range errs {
		parts = append(parts, er.Error())
	}
	return strings.Join(parts, "; ")
}

// Unwrap exposes the children so errors.Is and errors.AsType traverse
// every member of the group.
func (e *Error) Unwrap() []error {
	if len(e.extra) > 0 {
		return e.extra
	}
	return e.buf[:e.n]
}

// Format implements fmt.Formatter with the same verbs as Error: %v/%s
// render the single-line group, %+v renders each child with its own
// full formatting, %q quotes the single-line text.
func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if !s.Flag('+') {
			io.WriteString(s, e.Error())
			return
		}
		for i, er := range e.Unwrap() {
			if i > 0 {
				io.WriteString(s, "\n")
			}
			fmt.Fprintf(s, "%+v", er)
		}
	case 's':
		io.WriteString(s, e.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.Error())
	default:
		fmt.Fprintf(s, "%%!%c(%T=%v)", verb, e, e.Error())
	}
}

// LogValue implements slog.LogValuer: children are logged as a list,
// each as its own group when it exposes one and as its message
// otherwise; nesting is bounded to keep log records compact.
func (e *Error) LogValue() slog.Value {
	errs := e.Unwrap()
	children := make([]slog.Value, 0, len(errs))
	for _, er := range errs {
		children = append(children, errValue(er, 0))
	}
	return slog.GroupValue(
		slog.Int("count", len(children)),
		slog.Any("errors", children),
	)
}

const maxDepth = 3

func errValue(er error, depth int) slog.Value {
	if depth >= maxDepth {
		return slog.StringValue(er.Error())
	}
	switch e := er.(type) {
	case *Error:
		children := make([]slog.Value, 0, len(e.Unwrap()))
		for _, c := range e.Unwrap() {
			children = append(children, errValue(c, depth+1))
		}
		return slog.GroupValue(slog.Any("errors", children))
	case slog.LogValuer:
		return e.LogValue()
	default:
		return slog.StringValue(er.Error())
	}
}
