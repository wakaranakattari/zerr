package bench_test

import (
	stderrors "errors"
	"testing"

	cerrors "github.com/cockroachdb/errors"
	"github.com/samber/oops"
	"github.com/wakaranakattari/zerr"
	"github.com/wakaranakattari/zerr/join"
)

// This module exists to measure zerr head to head with the ecosystem
// on one wall clock: github.com/samber/oops and
// github.com/cockroachdb/errors, on this machine, same scenarios.
//
// Run: go test -bench . -benchmem ./...
//
// Numbers land in the README comparison table; re-measure before
// publishing new claims.

func BenchmarkZerrRoot(b *testing.B) {
	for i := 0; i < b.N; i++ {
		e := zerr.New("boom")
		_ = e
	}
}

func BenchmarkOopsRoot(b *testing.B) {
	for i := 0; i < b.N; i++ {
		e := oops.New("boom")
		_ = e
	}
}

func BenchmarkCockroachRoot(b *testing.B) {
	for i := 0; i < b.N; i++ {
		e := cerrors.Newf("boom")
		_ = e
	}
}

func BenchmarkZerrWrap(b *testing.B) {
	root := zerr.New("boom")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		e := zerr.Wrap(root, "load", "id", i, "user", "u42")
		_ = e
	}
}

func BenchmarkOopsWrap(b *testing.B) {
	root := oops.New("boom")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		e := oops.With("id", i, "user", "u42").Wrap(root)
		_ = e
	}
}

func BenchmarkCockroachWrap(b *testing.B) {
	root := cerrors.Newf("boom")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		e := cerrors.WrapWithDepth(0, root, "load")
		_ = e
	}
}

func BenchmarkZerrIs(b *testing.B) {
	const target = zerr.Code("not_found")
	err := zerr.New("boom")
	for i := 0; i < 10; i++ {
		err = zerr.Wrapf(err, "w", "%d", i)
	}
	err = zerr.WithCode(err, target, "root")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if !stderrors.Is(err, target) {
			b.Fatal("expected a match")
		}
	}
}

func BenchmarkCockroachIs(b *testing.B) {
	root := cerrors.Newf("boom")
	err := root
	for i := 0; i < 10; i++ {
		err = cerrors.Wrap(err, "level")
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if !stderrors.Is(err, root) {
			b.Fatal("expected a match")
		}
	}
}

// NOTE oops v0.19.x cannot be matched with errors.Is at all: its
// OopsError.Is() compares the error type head-on and panics on the
// embedded map. Its blessed classification path is oops.AsOops, used
// here.
func BenchmarkOopsIs(b *testing.B) {
	root := oops.New("boom")
	err := root
	for i := 0; i < 10; i++ {
		err = oops.Wrap(err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, ok := oops.AsOops(err); !ok {
			b.Fatal("expected a match")
		}
	}
}

func BenchmarkZerrJoin(b *testing.B) {
	a := zerr.New("a")
	c := zerr.New("b")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		e := join.Append(a, c)
		_ = e
	}
}

func BenchmarkStdlibJoin(b *testing.B) {
	a := stderrors.New("a")
	c := stderrors.New("b")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		e := stderrors.Join(a, c)
		_ = e
	}
}
