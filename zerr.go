package zerr

import (
	core "github.com/wakaranakattari/zerr/internal/core"
)

// Error is a single immutable node in an error chain. See the package
// documentation for how nodes compose; the implementation lives in
// internal/core.
type Error = core.Error

// Code is a machine-readable classification for errors. It implements
// error so codes can be used as sentinel targets with errors.Is:
//
//	var ErrNotFound = zerr.Code("not_found")
type Code = core.Code

// Attr is a single key-value attribute attached to an error node.
type Attr = core.Attr

// Node is an immutable snapshot of a single node in a chain, exposed
// for inspection and for adapters (log encoders, renderers,
// exporters) that live outside this package.
type Node = core.Node

// Nodes returns every node of err's chain, outermost first. A chain
// that is not a zerr chain is returned as a single message-only node.
func Nodes(err error) []Node { return core.Nodes(err) }

// New returns a root error with the given message.
func New(msg string) error { return core.New(msg) }

// Newf is New with formatting: Newf("open %q", f) is equivalent to
// New(fmt.Sprintf("open %q", f)).
func Newf(format string, args ...any) error { return core.Newf(format, args...) }

// Wrap adds context to err and returns a new node describing the
// operation op with zero or more attribute pairs. Wrap returns nil
// when err is nil, so chains can be built without guarding each call.
func Wrap(err error, op string, kv ...any) error { return core.Wrap(err, op, kv...) }

// Wrapf is Wrap with a formatted message.
func Wrapf(err error, op, format string, args ...any) error {
	return core.Wrapf(err, op, format, args...)
}

// WithCode is Wrap that additionally classifies the node with a code.
// Codes are matched with errors.Is(err, code) or zerr.Is(err, code).
func WithCode(err error, code Code, op string, kv ...any) error {
	return core.WithCode(err, code, op, kv...)
}

// Public wraps err and attaches a user-facing message: the text that is
// safe to hand to a client. The internal chain (op, msg, codes,
// attributes, private fields) stays on the server side.
func Public(err error, msg string, kv ...any) error {
	return core.Public(err, msg, kv...)
}

// Is reports whether err is classified with the given code anywhere in
// its chain.
func Is(err error, code Code) bool { return core.Is(err, code) }

// CodeOf returns the first code found while walking err's chain from
// the outermost node inward, or "" when the chain carries no code.
func CodeOf(err error) Code { return core.CodeOf(err) }

// Sec returns a private attribute, delivered only through Fields.
func Sec(key string, value any) Attr { return core.Sec(key, value) }

// Fields returns every attribute found in err's chain: public and
// private, outermost node first, first occurrence winning for
// duplicate keys.
func Fields(err error) []Attr { return core.Fields(err) }
