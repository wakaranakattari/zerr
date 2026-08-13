// Package zerr is a structured error library that fits the idiomatic
// Go error flow: errors keep the familiar "return error" shape, and
// every error is a plain value that can be wrapped, classified and
// logged without ceremony.
//
// The unit of everything is a node: one heap allocation carrying an
// operation name, an optional message, a user-facing public message,
// an optional machine-readable code, the wrapped cause and up to two
// inline attribute pairs. Nodes compose into chains through Wrap and
// friends; a chain behaves exactly like an ordinary Go error.
//
//   - Wrap/Wrapf add context; WithCode classifies with a Code.
//   - Code is itself an error, so errors.Is(err, code) matches at any
//     depth. Codes form families: Code("io.timeout") matches
//     Code("io"), and each node carries at most one code.
//   - Public attaches the text an API may return; Public() is the only
//     channel that crosses a boundary. Private attributes (Sec) are
//     delivered exclusively through Fields and never rendered by fmt.
//   - Errors are slog.LogValuers, so structured logging is free; the
//     zerr/zap submodule does the same for go.uber.org/zap.
//   - join groups independent failures, must turns panics into
//     boundary-safe errors, and httperr maps codes to HTTP statuses.
//
// Zero dependencies: the module uses only the standard library.
// Stack traces are opt-in: build with -tags herr_trace to record each
// node's call site inside the same single allocation.
package zerr
