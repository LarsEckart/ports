package render

import (
	"bytes"
	"testing"

	"github.com/LarsEckart/ports/scanner"
)

func TestCleanConfirmation(t *testing.T) {
	var out bytes.Buffer
	DisplayCleanConfirmation(&out, []scanner.PortInfo{{Port: 3000, ProcessName: "node", PID: 123}})
	DisplayCleanAborted(&out)
	want := "\nFound 1 orphaned/zombie process(es):\n  - :3000 — node (PID 123)\n\nAborted.\n"
	if out.String() != want {
		t.Fatalf("got %q, want %q", out.String(), want)
	}
}
