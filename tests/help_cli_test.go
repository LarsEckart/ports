package main_test

import (
	"strings"
	"testing"
)

func TestVersionFlag(t *testing.T) {
	stdout, stderr, exitCode := runCLI(t, "--version")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%s stdout=%s", exitCode, stderr, stdout)
	}
	if !strings.Contains(stdout, "ports version ") {
		t.Fatalf("expected version output, got:\n%s", stdout)
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", stderr)
	}
}

func TestRootHelpUsesOptionsHeading(t *testing.T) {
	stdout, stderr, exitCode := runCLI(t, "--help")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%s stdout=%s", exitCode, stderr, stdout)
	}
	if !strings.Contains(stdout, "OPTIONS:") {
		t.Fatalf("expected OPTIONS heading, got:\n%s", stdout)
	}
	if strings.Contains(stdout, "GLOBAL OPTIONS:") {
		t.Fatalf("did not expect GLOBAL OPTIONS heading, got:\n%s", stdout)
	}
}

func TestKillRejectsAllAsUsageError(t *testing.T) {
	stdout, stderr, exitCode := runCLI(t, "kill", "--all", "12345")
	if exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d, stderr=%s stdout=%s", exitCode, stderr, stdout)
	}
	if !strings.Contains(stderr, "Incorrect Usage: flag provided but not defined: -all") {
		t.Fatalf("expected usage error on stderr, got:\n%s", stderr)
	}
	if !strings.Contains(stdout, "ports kill [options]") {
		t.Fatalf("expected kill help on stdout, got:\n%s", stdout)
	}
}

func TestUnknownCommandReturnsUsageExitCode(t *testing.T) {
	stdout, stderr, exitCode := runCLI(t, "wat")
	if exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d, stderr=%s stdout=%s", exitCode, stderr, stdout)
	}
	if !strings.Contains(stderr, "unknown command or argument: wat") {
		t.Fatalf("expected unknown command error on stderr, got:\n%s", stderr)
	}
	if !strings.Contains(stdout, "ports [options]") {
		t.Fatalf("expected root help on stdout, got:\n%s", stdout)
	}
}

func TestKillWithoutArgsShowsUsageHelp(t *testing.T) {
	stdout, stderr, exitCode := runCLI(t, "kill")
	if exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d, stderr=%s stdout=%s", exitCode, stderr, stdout)
	}
	if !strings.Contains(stderr, "Incorrect Usage: usage: ports kill") {
		t.Fatalf("expected usage error on stderr, got:\n%s", stderr)
	}
	if !strings.Contains(stdout, "ports kill [options]") {
		t.Fatalf("expected kill help on stdout, got:\n%s", stdout)
	}
}

func TestKillMutuallyExclusiveFlagsShowUsageHelp(t *testing.T) {
	stdout, stderr, exitCode := runCLI(t, "kill", "--pid", "--port", "12345")
	if exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d, stderr=%s stdout=%s", exitCode, stderr, stdout)
	}
	if !strings.Contains(stderr, "Incorrect Usage: choose only one of --pid or --port") {
		t.Fatalf("expected usage error on stderr, got:\n%s", stderr)
	}
	if !strings.Contains(stdout, "ports kill [options]") {
		t.Fatalf("expected kill help on stdout, got:\n%s", stdout)
	}
}
