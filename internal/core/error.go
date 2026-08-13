package core

import (
	"fmt"
	"strings"
)

// Error is a single immutable node in an error chain.
//
// A node carries an operation name, an optional short message, an
// optional user-facing message, an optional machine-readable code, its
// cause (the wrapped error), and up to four attributes stored inline.
// Everything lives in one heap allocation: the same cost as errors.New.
type Error struct {
	op    string
	msg   string
	pub   string
	code  Code
	cause error

	attrs      [2]Attr
	nattrs     int
	attrsExtra []Attr

	frames      [8]uintptr
	nframes     int
	framesExtra []uintptr
}

// New returns a root error with the given message.
//
// A nil error is never returned for a non-empty message; each call to
// New creates a distinct node, exactly like errors.New.
func New(msg string) error {
	return newNode(nil, "", msg, "", "")
}

// Newf is New with formatting: Newf("open %q", f) is equivalent to
// New(fmt.Sprintf("open %q", f)). With no arguments Newf avoids the
// formatting entirely.
func Newf(format string, args ...any) error {
	if len(args) == 0 {
		return New(format)
	}
	return New(fmt.Sprintf(format, args...))
}

// Wrap adds context to err and returns a new node describing the
// operation op with zero or more attribute pairs.
//
// The node's message is empty; use Wrapf to attach a message.
// Attributes are alternating keys and values; keys must be strings.
// A zerr.Attr value may be passed directly instead of a pair, which is
// how private attributes (Sec) are attached.
//
// Wrap returns nil when err is nil, so chains can be built without
// guarding each call.
func Wrap(err error, op string, kv ...any) error {
	if err == nil {
		return nil
	}
	return newNode(err, op, "", "", "", kv...)
}

// Wrapf is Wrap with a formatted message: the formatted text is stored
// on the node and rendered in the chain.
func Wrapf(err error, op, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return newNode(err, op, fmt.Sprintf(format, args...), "", "", nil...)
}

// WithCode is Wrap that additionally classifies the node with a code.
// Codes are matched with errors.Is(err, code) or zerr.Is(err, code).
func WithCode(err error, code Code, op string, kv ...any) error {
	if err == nil {
		return nil
	}
	return newNode(err, op, "", code, "", kv...)
}

// Public wraps err and attaches a user-facing message: the text that is
// safe to hand to a client. The internal chain (op, msg, codes,
// attributes, private fields) stays on the server side and is only
// reachable through Error, Formats, Fields and LogValue.
//
// The public message is rendered in %+v output and returned by
// Public(); it never reaches logs. Use it at API boundaries:
//
//	return zerr.Public(realErr, "the account could not be loaded")
func Public(err error, msg string, kv ...any) error {
	if err == nil {
		return nil
	}
	return newNode(err, "", "", "", msg, kv...)
}

func newNode(cause error, op, msg string, code Code, pub string, kv ...any) error {
	e := &Error{op: op, msg: msg, pub: pub, code: code, cause: cause}
	e.attrs, e.nattrs, e.attrsExtra = parseKV(kv)
	e.captureStack()
	return e
}

// Error renders the chain as a compact single line, outermost cause
// first: "op3: op2: op1: root message". Neither codes nor attributes
// are rendered; use Format for %v variants and Fields for structured
// delivery.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	var b strings.Builder
	for node := e; node != nil; {
		if b.Len() > 0 {
			b.WriteString(": ")
		}
		if node.op != "" {
			b.WriteString(node.op)
			if node.msg != "" {
				b.WriteString(": ")
			}
		}
		if node.msg != "" {
			b.WriteString(node.msg)
		}
		cause, ok := node.cause.(*Error)
		if !ok {
			if c := node.cause; c != nil {
				if b.Len() > 0 {
					b.WriteString(": ")
				}
				b.WriteString(c.Error())
			}
			break
		}
		node = cause
	}
	return b.String()
}

// Unwrap returns the wrapped cause, or nil for a root node.
func (e *Error) Unwrap() error { return e.cause }

// Public returns the outermost user-facing message attached with
// Public, or "" when the chain carries none. It is the only text that
// should cross an API boundary; everything else stays internal.
func (e *Error) Public() string {
	for node := e; node != nil; {
		if node.pub != "" {
			return node.pub
		}
		cause, ok := node.cause.(*Error)
		if !ok {
			break
		}
		node = cause
	}
	return ""
}

// Is reports whether this node's code matches target when target is a
// Code. A node matches a Code target when its code is the target or
// belongs to the target's family: Code("io.timeout") matches
// Code("io"), so hierarchy-aware sentinels work with errors.Is at any
// depth through Unwrap.
func (e *Error) Is(target error) bool {
	c, ok := target.(Code)
	if !ok {
		return false
	}
	if e.code == c {
		return true
	}
	n, m := len(e.code), len(c)
	if m > 0 && n > m && e.code[m] == '.' {
		return e.code[:m] == c
	}
	return false
}

func (e *Error) attrsList() []Attr {
	if len(e.attrsExtra) > 0 {
		return e.attrsExtra
	}
	return e.attrs[:e.nattrs]
}

func (e *Error) framesPCs() []uintptr {
	if len(e.framesExtra) > 0 {
		return e.framesExtra
	}
	return e.frames[:e.nframes]
}

// Node is an immutable snapshot of a single node in a chain, exposed
// for inspection and for adapters (log encoders, renderers, exporters)
// that live outside this package. Nodes are ordered outermost first.
type Node struct {
	Op    string
	Msg   string
	Pub   string
	Code  Code
	Attrs []Attr
}

// Nodes returns every node of err's chain, outermost first. A chain
// that is not a zerr chain -- a foreign error, a join group -- is
// returned as a single message-only node, so callers can always assume
// at least one entry for a non-nil error.
func Nodes(err error) []Node {
	if err == nil {
		return nil
	}
	node, ok := err.(*Error)
	if !ok {
		return []Node{{Msg: err.Error()}}
	}
	out := make([]Node, 0, 4)
	for node != nil {
		attrs := node.attrsList()
		n := Node{Op: node.op, Msg: node.msg, Pub: node.pub, Code: node.code}
		if len(attrs) > 0 {
			n.Attrs = attrs
		}
		out = append(out, n)
		cause := node.cause
		node, ok = cause.(*Error)
		if !ok {
			if cause != nil {
				out = append(out, Node{Msg: cause.Error()})
			}
			break
		}
	}
	return out
}
