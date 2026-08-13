//go:build herr_trace

package core

import "runtime"

// captureStack records the call stack of the node's creation site.
// Up to eight frames fit in the node's own allocation; deeper stacks
// spill into a single backing slice.
//
// the skip of 2 accounts for runtime.Callers and captureStack itself
// (captureStack is not inlined, so the recorded frames are exactly
// [newNode, Wrap-or-user, user, ...]); formatters later drop every
// frame below the public call site, so the first recorded frame is
// always the caller of the public constructor regardless of inlining.
//
//go:noinline
func (e *Error) captureStack() {
	n := runtime.Callers(2, e.frames[:])
	e.nframes = n
	if n < len(e.frames) {
		return
	}
	buf := make([]uintptr, len(e.frames))
	for {
		n = runtime.Callers(2, buf)
		if n < len(buf) {
			e.framesExtra = buf[:n]
			return
		}
		buf = make([]uintptr, len(buf)*2)
	}
}
