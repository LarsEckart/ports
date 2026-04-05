package main_test

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestKillByPort(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 not available")
	}

	port := freePort(t)
	server := exec.CommandContext(t.Context(), python, "-m", "http.server", fmt.Sprintf("%d", port), "--bind", "127.0.0.1")
	if err := server.Start(); err != nil {
		t.Fatalf("failed to start http server: %v", err)
	}
	defer func() {
		if server.Process != nil {
			_ = server.Process.Kill()
			_, _ = server.Process.Wait()
		}
	}()

	waitForPort(t, port, 5*time.Second)

	stdout, stderr, exitCode := runCLI(t, "kill", fmt.Sprintf("%d", port))
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%s stdout=%s", exitCode, stderr, stdout)
	}
	if !strings.Contains(stdout, "Sent SIGTERM") {
		t.Fatalf("expected kill output, got:\n%s", stdout)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 100*time.Millisecond)
		if err != nil {
			return
		}
		_ = conn.Close()
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("port %d still accepting connections after kill", port)
}

func TestKillRequiresDisambiguationWhenPortAndPIDConflict(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 not available")
	}

	pidProcess := exec.CommandContext(t.Context(), "sleep", "30")
	if err := pidProcess.Start(); err != nil {
		t.Fatalf("failed to start pid target: %v", err)
	}
	defer func() {
		if pidProcess.Process != nil {
			_ = pidProcess.Process.Kill()
			_, _ = pidProcess.Process.Wait()
		}
	}()

	target := pidProcess.Process.Pid
	if target > 65535 {
		t.Skipf("pid %d is outside the TCP port range", target)
	}

	listener := exec.CommandContext(t.Context(), python, "-m", "http.server", fmt.Sprintf("%d", target), "--bind", "127.0.0.1")
	if err := listener.Start(); err != nil {
		t.Skipf("failed to bind listener on port %d: %v", target, err)
	}
	defer func() {
		if listener.Process != nil {
			_ = listener.Process.Kill()
			_, _ = listener.Process.Wait()
		}
	}()

	waitForPort(t, target, 5*time.Second)

	stdout, stderr, exitCode := runCLI(t, "kill", fmt.Sprintf("%d", target))
	if exitCode == 0 {
		t.Fatalf("expected ambiguity to fail, stdout=%s stderr=%s", stdout, stderr)
	}
	if !strings.Contains(stdout, "use --port or --pid") {
		t.Fatalf("expected disambiguation hint, got:\n%s", stdout)
	}
	if err := syscall.Kill(target, 0); err != nil {
		t.Fatalf("expected PID %d to still be alive after ambiguous kill attempt: %v", target, err)
	}
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", target), 100*time.Millisecond)
	if err != nil {
		t.Fatalf("expected listener on port %d to still be alive after ambiguous kill attempt: %v", target, err)
	}
	_ = conn.Close()
}
