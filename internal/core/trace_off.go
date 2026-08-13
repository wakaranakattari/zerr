//go:build !herr_trace

package core

// captureStack is a no-op in production builds. Stack traces are
// captured only when the package is compiled with the herr_trace build
// tag, keeping the default build at a single allocation per node.
func (e *Error) captureStack() {}
