package main_test

import (
	"fmt"
	"net"
	"strings"
	"testing"
)

func TestPortsAllShowsOpenListener(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer func() { _ = listener.Close() }()

	port := listener.Addr().(*net.TCPAddr).Port
	stdout, stderr, exitCode := runCLI(t, "--all")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%s", exitCode, stderr)
	}
	if !strings.Contains(stdout, fmt.Sprintf(":%d", port)) {
		t.Fatalf("expected output to contain port %d, got:\n%s", port, stdout)
	}
}

func TestPortDetailShowsPID(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer func() { _ = listener.Close() }()

	port := listener.Addr().(*net.TCPAddr).Port
	stdout, stderr, exitCode := runCLI(t, fmt.Sprintf("%d", port))
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%s", exitCode, stderr)
	}
	if !strings.Contains(stdout, fmt.Sprintf("Port :%d", port)) {
		t.Fatalf("expected detail output for port %d, got:\n%s", port, stdout)
	}
	if !strings.Contains(stdout, "PID") {
		t.Fatalf("expected PID in output, got:\n%s", stdout)
	}
}

func TestPSAllShowsTable(t *testing.T) {
	stdout, stderr, exitCode := runCLI(t, "ps", "--all")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%s", exitCode, stderr)
	}
	if !strings.Contains(stdout, "PROCESS") || !strings.Contains(stdout, "CPU%") {
		t.Fatalf("expected process table headers, got:\n%s", stdout)
	}
}

func TestPortsReportsPrimaryScanFailures(t *testing.T) {
	stdout, stderr, exitCode := runCLIWithEnv(t, []string{"PATH=/usr/bin:/bin"}, "--all")
	if exitCode == 0 {
		t.Fatalf("expected non-zero exit code when lsof is unavailable, stdout=%s stderr=%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "list listening ports") {
		t.Fatalf("expected scanner failure in stderr, got:\n%s", stderr)
	}
	if strings.Contains(stdout, "No active listening ports found") {
		t.Fatalf("expected scanner failure instead of empty-state output, got:\n%s", stdout)
	}
}
