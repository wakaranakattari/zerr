package core_test

import (
	"fmt"
	"strings"
	"testing"

	zerr "github.com/wakaranakattari/zerr/internal/core"
	"github.com/wakaranakattari/zerr/join"
)

func TestFieldsOrderOutermostFirst(t *testing.T) {
	err := zerr.Wrap(zerr.New("root"), "op")
	err = zerr.Wrap(err, "op2", "id", 1)
	fields := zerr.Fields(err)
	if len(fields) != 1 || fields[0].Key != "id" {
		t.Fatalf("Fields = %+v, want a single id attribute", fields)
	}
}

func TestFieldsDedupeKeepsOutermost(t *testing.T) {
	err := zerr.Wrap(zerr.New("root"), "inner", "id", 1)
	err = zerr.Wrap(err, "outer", "id", 2)
	fields := zerr.Fields(err)
	if len(fields) != 1 || fields[0].Value != 2 {
		t.Fatalf("Fields = %+v, want first occurrence (value 2) to win", fields)
	}
}

func TestFieldsIncludesPrivate(t *testing.T) {
	err := zerr.Wrap(zerr.New("root"), "login",
		"user", "alice",
		zerr.Sec("token", "secret"),
	)
	fields := zerr.Fields(err)
	if len(fields) != 2 {
		t.Fatalf("Fields = %+v, want 2 attributes", fields)
	}
	if !fields[1].Priv || fields[1].Key != "token" {
		t.Fatalf("Fields = %+v, want the private token attribute", fields)
	}
}

func TestPrivateNeverFormatted(t *testing.T) {
	err := zerr.Wrap(zerr.New("root"), "login", zerr.Sec("token", "secret"))
	for _, format := range []string{"%v", "%s", "%q", "%+v"} {
		out := fmt.Sprintf(format, err)
		if strings.Contains(out, "secret") || strings.Contains(out, "token") {
			t.Fatalf("%s leaked a private attribute: %q", format, out)
		}
	}
}

func TestFieldsAcrossJoin(t *testing.T) {
	a := zerr.Wrap(zerr.New("a"), "opA", "x", 1)
	b := zerr.Wrap(zerr.New("b"), "opB", "y", 2)
	fields := zerr.Fields(join.Join(a, b))
	if len(fields) != 2 {
		t.Fatalf("Fields = %+v, want attributes from both children", fields)
	}
}

func TestFieldsNilError(t *testing.T) {
	if fields := zerr.Fields(nil); len(fields) != 0 {
		t.Fatalf("Fields(nil) = %+v, want empty", fields)
	}
}
