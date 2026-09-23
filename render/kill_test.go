package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/LarsEckart/ports/scanner"
)

func TestKillMessages(t *testing.T) {
	var out bytes.Buffer
	DisplayInvalidKillArgument(&out, "oops")
	DisplayAmbiguousKillTarget(&out, 3000)
	DisplayMissingKillTarget(&out, 3000, false, true)
	DisplayMissingKillTarget(&out, 3000, true, false)
	DisplayMissingKillTarget(&out, 3000, false, false)
	DisplayMissingKillTarget(&out, 70000, false, false)
	text := out.String()
	for _, want := range []string{
		`"oops" is not a valid port or PID`,
		"3000 matches both a listening port and a PID; use --port or --pid",
		"No listener on :3000",
		"No process with PID 3000",
		"No listener on :3000 and no process with PID 3000",
		"No process with PID 70000",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in %q", want, text)
		}
	}
}

func TestKillResultLabelsAndSignals(t *testing.T) {
	var out bytes.Buffer
	port := &scanner.KillTarget{PID: 123, Port: 3000, Via: "port", Info: &scanner.PortInfo{ProcessName: "node"}}
	DisplayKilling(&out, port)
	DisplayKillResult(&out, port, false, true)
	DisplayKillResult(&out, port, true, false)
	pid := &scanner.KillTarget{PID: 456}
	DisplayKilling(&out, pid)
	text := out.String()
	for _, want := range []string{
		"Killing :3000 — node (PID 123)",
		"Sent SIGTERM to :3000 — node (PID 123)",
		"Failed to send SIGKILL to :3000 — node (PID 123)",
		"Killing PID 456",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in %q", want, text)
		}
	}
}
