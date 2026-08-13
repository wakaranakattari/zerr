package core_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	zerr "github.com/wakaranakattari/zerr/internal/core"
	"github.com/wakaranakattari/zerr/join"
)

func TestFormatCompact(t *testing.T) {
	err := zerr.Wrap(zerr.Wrap(zerr.New("boom"), "load"), "handle")
	for _, format := range []string{"%v", "%s"} {
		if got, want := fmt.Sprintf(format, err), "handle: load: boom"; got != want {
			t.Fatalf("%s = %q, want %q", format, got, want)
		}
	}
}

func TestFormatQuoted(t *testing.T) {
	err := zerr.Wrapf(zerr.New("boom"), "load", "id %d", 42)
	if got, want := fmt.Sprintf("%q", err), `"load: id 42: boom"`; got != want {
		t.Fatalf("got %q, want %s", got, want)
	}
}

func TestFormatVerbose(t *testing.T) {
	err := zerr.WithCode(zerr.Wrap(zerr.New("boom"), "db"), zerr.Code("not_found"), "load", "id", 7)
	out := fmt.Sprintf("%+v", err)
	for _, want := range []string{"load", "code: not_found", "id: 7", "db", "boom"} {
		if !strings.Contains(out, want) {
			t.Fatalf("verbose %q missing %q", out, want)
		}
	}
}

func TestFormatVerbosePublic(t *testing.T) {
	err := zerr.Public(zerr.Wrap(zerr.New("no rows"), "db"), "account unavailable")
	out := fmt.Sprintf("%+v", err)
	if !strings.Contains(out, "public: account unavailable") {
		t.Fatalf("verbose %q must render the public message, got %q", out, err)
	}
	if got := fmt.Sprintf("%v", err); strings.Contains(got, "account unavailable") {
		t.Fatalf("compact %q must not leak the public message", got)
	}
}

func TestFormatVerboseForeignTail(t *testing.T) {
	err := zerr.Wrap(errors.New("foreign"), "op")
	out := fmt.Sprintf("%+v", err)
	if !strings.Contains(out, "cause: foreign") {
		t.Fatalf("verbose %q must mark a foreign cause", out)
	}
}

func TestFormatVerboseOmittedWhenNoAttrs(t *testing.T) {
	err := zerr.New("boom")
	out := fmt.Sprintf("%+v", err)
	if strings.Contains(out, "code:") {
		t.Fatalf("must not print an empty code: %s", out)
	}
}

func TestFormatUnknownVerb(t *testing.T) {
	err := zerr.New("boom")
	out := fmt.Sprintf("%d", err)
	if out == "boom" || strings.Contains(out, "boom") == false {
		// %d must produce the fmt error placeholder, not crash.
		if !strings.HasPrefix(out, "%!") {
			t.Fatalf("unexpected output for unknown verb: %q", out)
		}
	}
}

func TestFormatNilNode(t *testing.T) {
	var err *zerr.Error
	if got := err.Error(); got != "" {
		t.Fatalf("nil *Error rendered as %q", got)
	}
}

func TestJoinFormatCompact(t *testing.T) {
	joined := join.Join(zerr.New("a"), zerr.New("b"))
	if got, want := fmt.Sprintf("%v", joined), "a; b"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestJoinFormatVerbose(t *testing.T) {
	joined := join.Join(
		zerr.Wrap(zerr.New("a"), "opA", "x", 1),
		zerr.New("b"),
	)
	out := fmt.Sprintf("%+v", joined)
	if !strings.Contains(out, "opA") || !strings.Contains(out, "x: 1") || !strings.Contains(out, "b") {
		t.Fatalf("joined children missing from %q", out)
	}
}

func TestJoinWrappedInChain(t *testing.T) {
	err := zerr.Wrap(join.Join(zerr.New("a"), zerr.New("b")), "op")
	if got, want := err.Error(), "op: a; b"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}
