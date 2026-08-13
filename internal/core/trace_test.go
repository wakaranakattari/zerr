//go:build herr_trace

package core_test

import (
	"fmt"
	"strings"
	"testing"

	zerr "github.com/wakaranakattari/zerr/internal/core"
)

func TestStackCapturedInTraceBuild(t *testing.T) {
	err := zerr.Wrap(zerr.New("boom"), "load")
	out := fmt.Sprintf("%+v", err)

	if !strings.Contains(out, "core_test.TestStackCapturedInTraceBuild") {
		t.Fatalf("missing the caller's frame: %q", out)
	}
	if strings.Contains(out, "zerr.Wrap") {
		t.Fatalf("leaked an internal frame: %q", out)
	}
}

func TestStackPerNode(t *testing.T) {
	root := zerr.Wrap(zerr.New("boom"), "inner")
	err := zerr.Wrap(root, "outer")
	out := fmt.Sprintf("%+v", err)

	// Every node captures its own stack: three call sites, one per node.
	if got := strings.Count(out, "core_test.TestStackPerNode"); got != 3 {
		t.Fatalf("found %d stack sites, want 3 (one per node): %q", got, out)
	}
}
