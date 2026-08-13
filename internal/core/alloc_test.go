//go:build !herr_trace

package core_test

import (
	"testing"

	zerr "github.com/wakaranakattari/zerr/internal/core"
)

var allocSink error

func TestAllocsWrap(t *testing.T) {
	root := zerr.New("boom")
	assertAllocs(t, 1, func() { allocSink = zerr.Wrap(root, "op") })
}

func TestAllocsWrapTwoAttrsStayInlined(t *testing.T) {
	root := zerr.New("boom")
	assertAllocs(t, 1, func() {
		allocSink = zerr.Wrap(root, "op", "a", 1, "b", 2)
	})
}

func TestAllocsWrapThreeAttrsSpill(t *testing.T) {
	root := zerr.New("boom")
	assertAllocs(t, 2, func() {
		allocSink = zerr.Wrap(root, "op", "a", 1, "b", 2, "c", 3)
	})
}

func TestAllocsWithCode(t *testing.T) {
	root := zerr.New("boom")
	assertAllocs(t, 1, func() {
		allocSink = zerr.WithCode(root, zerr.Code("x"), "op")
	})
}

func TestAllocsNew(t *testing.T) {
	assertAllocs(t, 1, func() { allocSink = zerr.New("boom") })
}

func TestAllocsPublic(t *testing.T) {
	root := zerr.New("boom")
	assertAllocs(t, 1, func() { allocSink = zerr.Public(root, "user sees this") })
}

func TestAllocsWrapNil(t *testing.T) {
	assertAllocs(t, 0, func() { allocSink = zerr.Wrap(nil, "op") })
}

func assertAllocs(t *testing.T, want float64, fn func()) {
	t.Helper()
	if got := testing.AllocsPerRun(1000, fn); got != want {
		t.Fatalf("allocations = %v, want %v", got, want)
	}
}
