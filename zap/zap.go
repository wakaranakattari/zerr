// Package zap adapts zerr errors to go.uber.org/zap structured
// logging. One call, no glue code:
//
//	logger.Error("request failed", zapzap.Field(err))
//
// The error is encoded as a nested object with the same shape as the
// log/slog output: the outermost node becomes a group of op, msg and
// code attributes, its public attributes become siblings, and the
// cause is folded in as a nested "cause" group, bounded to three
// levels. Private attributes (zerr.Sec) are never encoded here --
// deliver them explicitly with zerr.Fields.
package zap

import (
	"strings"

	"github.com/wakaranakattari/zerr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// maxDepth mirrors the cause-depth bound of the slog encoder.
const maxDepth = 3

// Field encodes err as a zap object field. A nil error encodes to an
// empty object; a foreign or joined error collapses to its message.
func Field(err error) zap.Field {
	return zap.Object("err", nodesMarshaler{nodes: zerr.Nodes(err)})
}

type nodesMarshaler struct {
	nodes []zerr.Node
	depth int
}

func (m nodesMarshaler) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	writeNodes(enc, m.nodes, m.depth)
	return nil
}

// writeNodes folds nodes outermost-first into nested cause groups.
func writeNodes(enc zapcore.ObjectEncoder, nodes []zerr.Node, depth int) {
	if len(nodes) == 0 {
		return
	}
	head := nodes[0]
	if head.Op != "" {
		enc.AddString("op", head.Op)
	}
	if head.Msg != "" {
		enc.AddString("msg", head.Msg)
	}
	if head.Code != "" {
		enc.AddString("code", string(head.Code))
	}
	for _, a := range head.Attrs {
		if a.Priv {
			continue
		}
		enc.AddReflected(a.Key, a.Value)
	}
	if tail := nodes[1:]; len(tail) > 0 {
		if depth >= maxDepth {
			enc.AddString("cause", compact(tail))
			return
		}
		enc.AddObject("cause", nodesMarshaler{nodes: tail, depth: depth + 1})
	}
}

// compact renders the remaining chain the way Error() does, used once
// the cause-depth bound is reached.
func compact(nodes []zerr.Node) string {
	var b strings.Builder
	for i, n := range nodes {
		if i > 0 {
			b.WriteString(": ")
		}
		if n.Op != "" {
			b.WriteString(n.Op)
			if n.Msg != "" {
				b.WriteString(": ")
			}
		}
		b.WriteString(n.Msg)
	}
	return b.String()
}
