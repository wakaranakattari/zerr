//go:build !herr_trace

package zerr_test

import (
	"fmt"
	"testing"

	"github.com/wakaranakattari/zerr"
)

// Formatting and allocation guarantees hold for the production build;
// under herr_trace the verbose format gains stack lines and stack
// capture adds runtime work, both covered by internal/core tests.

func TestFacadeFormatting(t *testing.T) {
	err := zerr.Wrap(zerr.New("boom"), "op", "id", 7)
	if got, want := fmt.Sprintf("%+v", err), "op\n  id: 7\nboom"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestFacadeAllocs(t *testing.T) {
	root := zerr.New("boom")
	if got := testing.AllocsPerRun(1000, func() { apiSink = zerr.Wrap(root, "op") }); got != 1 {
		t.Fatalf("Wrap allocations = %v, want 1", got)
	}
}
