package zerr_test

import (
	"errors"
	"testing"

	"github.com/wakaranakattari/zerr"
)

var apiSink error

func TestFacadeChain(t *testing.T) {
	err := zerr.Wrap(zerr.WithCode(zerr.New("boom"), zerr.Code("x"), "store"), "api")
	if got, want := err.Error(), "api: store: boom"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if !zerr.Is(err, zerr.Code("x")) {
		t.Fatal("Is through the facade must match a deep code")
	}
	if got := zerr.CodeOf(err); got != zerr.Code("x") {
		t.Fatalf("CodeOf = %q, want %q", got, zerr.Code("x"))
	}
	var node *zerr.Error
	if !errors.As(err, &node) {
		t.Fatal("errors.As must see the facade alias as *zerr.Error")
	}
	if node == nil {
		t.Fatal("errors.As resolved to nil")
	}
}
