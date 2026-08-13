package must_test

import (
	"errors"
	"testing"

	"github.com/wakaranakattari/zerr"
	"github.com/wakaranakattari/zerr/must"
)

func TestMustReturnsValue(t *testing.T) {
	if got := must.Must(load(1)); got != "v1" {
		t.Fatalf("Must = %q, want v1", got)
	}
}

func TestMustPanicsWithError(t *testing.T) {
	var got any
	func() {
		defer func() { got = recover() }()
		_, err := load(2)
		must.Must("", err)
	}()
	if got == nil {
		t.Fatal("Must did not panic")
	}
	if _, ok := got.(error); !ok {
		t.Fatalf("Must panicked with %T, want an error value", got)
	}
}

func TestMustErr(t *testing.T) {
	var got any
	func() {
		defer func() { got = recover() }()
		_, err := load(2)
		must.MustErr(err)
	}()
	if got == nil {
		t.Fatal("MustErr did not panic")
	}
}

func TestCatchRestoresTypedError(t *testing.T) {
	code := zerr.Code("oops")
	cause := zerr.WithCode(zerr.New("boom"), code, "load")

	err := func() (err error) {
		defer must.Catch(&err)
		must.MustErr(cause)
		return nil
	}()

	if !errors.Is(err, code) {
		t.Fatalf("Catch restored the wrong error: %v", err)
	}
}

func TestCatchNonErrorPanic(t *testing.T) {
	err := func() (err error) {
		defer must.Catch(&err)
		panic("boom")
	}()
	if err == nil || err.Error() != "panic: boom" {
		t.Fatalf("Catch wrapped the panic as %q, want %q", err, "panic: boom")
	}
}

func TestCatchNoPanicDoesNothing(t *testing.T) {
	result := func() (err error) {
		defer must.Catch(&err)
		return zerr.New("fine")
	}()
	if result.Error() != "fine" {
		t.Fatalf("Catch clobbered a clean return: %v", result)
	}
}

func TestCatchRepanicsOnNilTarget(t *testing.T) {
	var got any
	func() {
		defer func() { got = recover() }()
		func() {
			defer must.Catch(nil)
			panic("once")
		}()
	}()
	if got == nil || got.(string) != "once" {
		t.Fatalf("Catch(nil) must re-panic, got %v", got)
	}
}

func load(id int) (string, error) {
	if id == 2 {
		return "", zerr.New("missing")
	}
	return "v1", nil
}
