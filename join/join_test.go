package join_test

import (
	"errors"
	"testing"

	"github.com/wakaranakattari/zerr"
	"github.com/wakaranakattari/zerr/join"
)

func TestJoinFiltersNil(t *testing.T) {
	if err := join.Join(nil, nil); err != nil {
		t.Fatalf("Join of nils = %v, want nil", err)
	}
	one := zerr.New("one")
	if err := join.Join(nil, one, nil); !errors.Is(err, one) {
		t.Fatal("Join must keep the non-nil member reachable")
	}
}

func TestJoinSinglePreservesIdentity(t *testing.T) {
	one := zerr.New("one")
	joined := join.Join(nil, one)
	if joined != one {
		t.Fatal("Join with a single non-nil error must return it unchanged")
	}
}

func TestJoinUnwrapAll(t *testing.T) {
	a, b := zerr.New("a"), zerr.New("b")
	joined := join.Join(a, b)
	children := joined.(interface{ Unwrap() []error }).Unwrap()
	if len(children) != 2 || children[0] != a || children[1] != b {
		t.Fatalf("Unwrap = %v, want [a b]", children)
	}
}

func TestJoinMatchesEachChild(t *testing.T) {
	a := zerr.WithCode(zerr.New("a"), zerr.Code("code_a"), "opA")
	b := zerr.New("b")
	joined := join.Join(a, b)
	if !errors.Is(joined, zerr.Code("code_a")) {
		t.Fatal("errors.Is must reach codes inside joined children")
	}
	if !errors.Is(joined, a) {
		t.Fatal("errors.Is must reach joined children")
	}
}

func TestJoinMoreThanFourChildren(t *testing.T) {
	errs := make([]error, 8)
	for i := range errs {
		errs[i] = zerr.Newf("e%d", i)
	}
	joined := join.Join(errs...)
	for i, e := range errs {
		if !errors.Is(joined, e) {
			t.Fatalf("child %d not reachable through the group", i)
		}
	}
	if got, want := joined.Error(), "e0; e1; e2; e3; e4; e5; e6; e7"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestAppendAccumulates(t *testing.T) {
	children := make([]error, 5)
	for i := range children {
		children[i] = zerr.Newf("e%d", i)
	}
	var err error
	for _, e := range children {
		err = join.Append(err, e)
	}
	if err == nil {
		t.Fatal("Append returned nil after accumulating errors")
	}
	for _, e := range children {
		if !errors.Is(err, e) {
			t.Fatalf("%q missing from appended group", e)
		}
	}
}

func TestAppendNilSafe(t *testing.T) {
	one := zerr.New("one")
	if err := join.Append(nil, nil); err != nil {
		t.Fatalf("Append(nil, nil) = %v, want nil", err)
	}
	if err := join.Append(one, nil); !errors.Is(err, one) {
		t.Fatal("Append with nil extras must preserve the original")
	}
	if err := join.Append(nil, one); err != one {
		t.Fatal("Append(nil, one) must return one unchanged")
	}
}

func TestFieldsAcrossNestedJoin(t *testing.T) {
	a := zerr.Wrap(zerr.New("a"), "opA", "x", 1)
	b := zerr.Wrap(zerr.New("b"), "opB", "y", 2)
	err := zerr.Wrap(join.Join(a, b), "outer")
	if got := zerr.Fields(err); len(got) != 2 {
		t.Fatalf("Fields = %+v, want attributes from both children", got)
	}
}

func TestJoinAllocs(t *testing.T) {
	a, b := zerr.New("a"), zerr.New("b")
	var sink error
	assertJoinAllocs(t, 1, func() { sink = join.Join(a, b) })
	_ = sink
}

func assertJoinAllocs(t *testing.T, want float64, fn func()) {
	t.Helper()
	if got := testing.AllocsPerRun(1000, fn); got != want {
		t.Fatalf("allocations = %v, want %v", got, want)
	}
}

func BenchmarkJoin(b *testing.B) {
	a, c := zerr.New("a"), zerr.New("c")
	var sink error
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sink = join.Join(a, c)
	}
	_ = sink
}
