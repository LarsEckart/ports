package scanner

import (
	"path/filepath"
	"testing"
)

func TestParseCWDOutput(t *testing.T) {
	output := `COMMAND   PID USER   FD   TYPE DEVICE SIZE/OFF    NODE NAME
node    48273 lars  cwd    DIR   1,15      896 2478558 /Users/lars/GitHub/meiner
node    48315 lars  cwd    DIR   1,15      896 2478558 /Users/lars/GitHub/meiner
`

	got := parseCWDOutput(output)
	want := filepath.Clean("/Users/lars/GitHub/meiner")

	if got[48273] != want {
		t.Fatalf("expected PID 48273 cwd %q, got %q", want, got[48273])
	}
	if got[48315] != want {
		t.Fatalf("expected PID 48315 cwd %q, got %q", want, got[48315])
	}
}

func TestParseCWDOutputIgnoresNoise(t *testing.T) {
	output := `COMMAND   PID USER   FD   TYPE DEVICE SIZE/OFF NODE NAME
bad line
node    nope lars  cwd    DIR   1,15      896    2 /tmp/project
node    42 lars  cwd    DIR   1,15      896    2 /tmp/project with spaces
`

	got := parseCWDOutput(output)
	if _, ok := got[0]; ok {
		t.Fatal("did not expect invalid PID entry")
	}
	if got[42] != "/tmp/project with spaces" {
		t.Fatalf("expected PID 42 cwd with spaces, got %q", got[42])
	}
}
