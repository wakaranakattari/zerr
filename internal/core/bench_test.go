package core_test

import (
	"errors"
	"fmt"
	"testing"

	zerr "github.com/wakaranakattari/zerr/internal/core"
)

var (
	benchSinkError error
	benchSinkBool  bool
	benchSinkStr   string
)

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSinkError = zerr.New("boom")
	}
}

func BenchmarkErrorsNew(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSinkError = errors.New("boom")
	}
}

func BenchmarkWrap(b *testing.B) {
	root := zerr.New("boom")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSinkError = zerr.Wrap(root, "db")
	}
}

func BenchmarkWrapAttrs4(b *testing.B) {
	root := zerr.New("boom")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSinkError = zerr.Wrap(root, "db", "a", 1, "b", 2, "c", 3, "d", 4)
	}
}

func BenchmarkWrapf(b *testing.B) {
	root := zerr.New("boom")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSinkError = zerr.Wrapf(root, "db", "load item %d", i)
	}
}

func BenchmarkFmtErrorf(b *testing.B) {
	root := errors.New("boom")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSinkError = fmt.Errorf("db: %w", root)
	}
}

func BenchmarkIsDeepChain(b *testing.B) {
	const code = zerr.Code("not_found")
	err := zerr.WithCode(zerr.New("root"), code, "store")
	for i := 0; i < 9; i++ {
		err = zerr.Wrap(err, "layer")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSinkBool = zerr.Is(err, code)
	}
}

func BenchmarkIsDeepChainStdlib(b *testing.B) {
	const sentinel = stdSentinel("not_found")
	chain := fmt.Errorf("store: %w", sentinel)
	for i := 0; i < 9; i++ {
		chain = fmt.Errorf("layer: %w", chain)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSinkBool = errors.Is(chain, sentinel)
	}
}

type stdSentinel string

func (s stdSentinel) Error() string { return string(s) }

func BenchmarkErrorRender(b *testing.B) {
	err := zerr.New("root")
	for i := 0; i < 9; i++ {
		err = zerr.Wrap(err, "layer")
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchSinkStr = err.Error()
	}
}
