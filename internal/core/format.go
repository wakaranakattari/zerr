package core

import (
	"fmt"
	"io"
	"runtime"
	"strings"
)

// Format implements fmt.Formatter. Supported verbs:
//
//	%v, %s  compact single-line chain (same text as Error)
//	%q      the chain, quoted
//	%+v     multiline rendering: every node on its own line with its
//	        public attributes and, in herr_trace builds, its stack
//
// Private attributes are never rendered, including in %+v.
func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			e.fmtVerbose(s)
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, e.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.Error())
	default:
		fmt.Fprintf(s, "%%!%c(%T=%v)", verb, e, e.Error())
	}
}

func (e *Error) fmtVerbose(w io.Writer) {
	node := e
	var tail error
	first := true
	for {
		if !first {
			io.WriteString(w, "\n")
		}
		first = false
		if node.op != "" {
			io.WriteString(w, node.op)
			if node.msg != "" {
				io.WriteString(w, ": ")
			}
		}
		io.WriteString(w, node.msg)
		if node.pub != "" {
			fmt.Fprintf(w, "\n  public: %s", node.pub)
		}
		if node.code != "" {
			fmt.Fprintf(w, "\n  code: %s", node.code)
		}
		for _, a := range node.attrsList() {
			if a.Priv {
				continue
			}
			fmt.Fprintf(w, "\n  %s: %v", a.Key, a.Value)
		}
		for _, l := range renderFrames(node) {
			fmt.Fprintf(w, "\n  at %s", l)
		}
		cause, ok := node.cause.(*Error)
		if !ok {
			tail = node.cause
			break
		}
		node = cause
	}
	if tail != nil {
		fmt.Fprintf(w, "\n  cause: %s", tail.Error())
	}
}

func renderFrames(e *Error) []string {
	pcs := e.framesPCs()
	if len(pcs) == 0 {
		return nil
	}
	frames := runtime.CallersFrames(pcs)
	out := make([]string, 0, len(pcs))
	for {
		f, more := frames.Next()
		if f.Function != "" && !strings.HasPrefix(f.Function, pkgPrefix+".") {
			out = append(out, fmt.Sprintf("%s:%d: %s", f.File, f.Line, f.Function))
		}
		if !more {
			break
		}
	}
	return out
}

// pkgPrefix is the runtime name prefix of this package (for example
// "github.com/wakaranakattari/zerr"), derived from a known non-inlined symbol
// so internal frames can be filtered from stack output.
var pkgPrefix = zerrAnchor()

//go:noinline
func zerrAnchor() string {
	pcs := make([]uintptr, 1)
	n := runtime.Callers(1, pcs)
	if n == 0 {
		return ""
	}
	name := runtime.FuncForPC(pcs[0]).Name()
	if i := strings.LastIndex(name, "."); i >= 0 {
		return name[:i]
	}
	return ""
}
