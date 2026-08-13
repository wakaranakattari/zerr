package core

import (
	"log/slog"

	"github.com/wakaranakattari/zerr/internal/multierror"
)

// LogValue implements slog.LogValuer so zerr errors log as structured
// groups: the operation, message and code become group attributes and
// every public attribute becomes a sibling field. An *Error cause is
// nested as a "cause" group (bounded depth); foreign errors and
// multi-error groups are collapsed to their message.
//
// Private attributes are intentionally absent: they are delivered with
// Fields instead.
func (e *Error) LogValue() slog.Value {
	return slog.GroupValue(e.groupAttrs(0)...)
}

const maxCauseDepth = 3

func (e *Error) groupAttrs(depth int) []slog.Attr {
	out := make([]slog.Attr, 0, 6)
	if e.op != "" {
		out = append(out, slog.String("op", e.op))
	}
	if e.msg != "" {
		out = append(out, slog.String("msg", e.msg))
	}
	if e.code != "" {
		out = append(out, slog.String("code", string(e.code)))
	}
	for _, a := range e.attrsList() {
		if a.Priv {
			continue
		}
		out = append(out, slog.Any(a.Key, a.Value))
	}
	if e.cause != nil {
		out = append(out, slog.Any("cause", errorValue(e.cause, depth+1)))
	}
	return out
}

// LogValue implements slog.LogValuer for join and must groups...
func errorValue(er error, depth int) slog.Value {
	switch e := er.(type) {
	case *Error:
		if depth >= maxCauseDepth {
			return slog.StringValue(e.Error())
		}
		return slog.GroupValue(e.groupAttrs(depth)...)
	case *multierror.Error:
		children := make([]slog.Value, 0, len(e.Unwrap()))
		for _, c := range e.Unwrap() {
			children = append(children, errorValue(c, depth+1))
		}
		return slog.GroupValue(slog.Any("errors", children))
	case slog.LogValuer:
		return e.LogValue()
	default:
		return slog.StringValue(er.Error())
	}
}
