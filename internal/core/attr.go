package core

import (
	"fmt"

	"github.com/wakaranakattari/zerr/internal/multierror"
)

// Attr is a single key-value attribute attached to an error node.
// Values are stored as-is; no formatting or copying takes place at
// creation time.
type Attr struct {
	Key   string
	Value any
	Priv  bool
}

// Sec returns a private attribute. Private attributes are never
// rendered by Format, %v, %+v or LogValue; the only way to obtain
// them is Fields, which exists precisely for internal log delivery.
//
// It is used inside Wrap's kv arguments:
//
//	err := zerr.Wrap(err, "login", "user", user, zerr.Sec("token", t))
func Sec(key string, value any) Attr {
	return Attr{Key: key, Value: value, Priv: true}
}

// Fields returns every attribute found in err's chain: public and
// private, outermost node first. Attributes with duplicate keys keep
// only their first occurrence, so the most recently wrapped value wins.
func Fields(err error) []Attr {
	out := make([]Attr, 0, 4)
	seen := make(map[string]struct{}, 8)
	var walk func(error)
	walk = func(er error) {
		switch e := er.(type) {
		case *Error:
			for _, a := range e.attrsList() {
				if _, dup := seen[a.Key]; dup {
					continue
				}
				seen[a.Key] = struct{}{}
				out = append(out, a)
			}
			if e.cause != nil {
				walk(e.cause)
			}
		case *multierror.Error:
			for _, c := range e.Unwrap() {
				walk(c)
			}
		}
	}
	walk(err)
	return out
}

// parseKV turns a flat kv argument list into an inline attribute buffer.
// A list of up to two pairs (or Attr values) fits in the node's own
// allocation; longer lists allocate a single backing slice.
func parseKV(kv []any) (inline [2]Attr, n int, extra []Attr) {
	total := countKV(kv)
	switch {
	case total == 0:
		return inline, 0, nil
	case total <= len(inline):
		fillKV(inline[:total], kv)
		return inline, total, nil
	default:
		extra = make([]Attr, total)
		fillKV(extra, kv)
		return inline, 0, extra
	}
}

func countKV(kv []any) (n int) {
	for i := 0; i < len(kv); {
		switch k := kv[i].(type) {
		case Attr:
			i++
		case string:
			i += 2
			if i > len(kv) {
				panic("zerr: odd number of attribute key/value pairs")
			}
		default:
			panic(fmt.Sprintf("zerr: attribute key must be string or zerr.Attr, got %T", k))
		}
		n++
	}
	return
}

func fillKV(dst []Attr, kv []any) {
	i := 0
	for pos := 0; pos < len(dst); pos++ {
		if a, ok := kv[i].(Attr); ok {
			dst[pos] = a
			i++
			continue
		}
		key := kv[i].(string)
		dst[pos] = Attr{Key: key, Value: kv[i+1]}
		i += 2
	}
}
