package core_test

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	zerr "github.com/wakaranakattari/zerr/internal/core"
	"github.com/wakaranakattari/zerr/join"
)

func TestLogValueStructured(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	err := zerr.WithCode(
		zerr.Wrap(zerr.New("boom"), "db", "id", 7),
		zerr.Code("not_found"), "load",
		"user", "alice",
	)
	logger.Info("operation failed", "err", err)

	out := buf.String()
	for _, want := range []string{
		"code=not_found",
		"user=alice",
		"cause",
		"db",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("slog output %q missing %q", out, want)
		}
	}
}

func TestLogValueOmitsPrivate(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	err := zerr.Wrap(zerr.New("boom"), "login",
		"user", "alice",
		zerr.Sec("token", "secret"),
	)
	logger.Info("login failed", "err", err)

	out := buf.String()
	if strings.Contains(out, "secret") {
		t.Fatalf("slog output leaked a private attribute: %q", out)
	}
}

func TestLogValueDepthBound(t *testing.T) {
	err := zerr.New("root")
	for i := 0; i < 10; i++ {
		err = zerr.Wrap(err, "level")
	}
	// The cause group is bounded, and the value must still render.
	var buf bytes.Buffer
	slog.New(slog.NewTextHandler(&buf, nil)).Info("deep", "err", err)
	if !strings.Contains(buf.String(), "level") {
		t.Fatalf("deep chain did not render: %q", buf.String())
	}
}

func TestJoinLogValue(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	err := join.Join(
		zerr.Wrap(zerr.New("a"), "opA", "x", 1),
		zerr.New("b"),
	)
	logger.Info("batch failed", "err", err)

	out := buf.String()
	for _, want := range []string{"count=2", "opA", "b"} {
		if !strings.Contains(out, want) {
			t.Fatalf("slog output %q missing %q", out, want)
		}
	}
}
