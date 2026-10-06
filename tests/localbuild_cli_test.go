package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestPortsShowsLocallyBuiltBinary covers a compiled app, such as a Roc or Go
// build, that serves a port. Its name carries no language hint, so the only
// signal is that the binary sits inside its own project.
func TestPortsShowsLocallyBuiltBinary(t *testing.T) {
	port := startLocalApp(t)

	stdout, stderr, exitCode := runCLI(t)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%s", exitCode, stderr)
	}
	if !strings.Contains(stdout, ":"+strconv.Itoa(port)) {
		t.Fatalf("expected default view to show port %d, got:\n%s", port, stdout)
	}
	if !strings.Contains(stdout, "Roc") {
		t.Fatalf("expected Roc framework in output, got:\n%s", stdout)
	}
}

// startLocalApp builds a Roc-style project with a compiled app in build/ and
// returns the port that app listens on.
func startLocalApp(t *testing.T) int {
	t.Helper()

	project := t.TempDir()
	for _, dir := range []string{".git", "app", "build"} {
		if err := os.MkdirAll(filepath.Join(project, dir), 0o755); err != nil {
			t.Fatalf("failed to create %s: %v", dir, err)
		}
	}
	if err := os.WriteFile(filepath.Join(project, "app", "main.roc"), []byte("app \"localapp\"\n"), 0o644); err != nil {
		t.Fatalf("failed to write app/main.roc: %v", err)
	}

	build := exec.CommandContext(t.Context(), "go", "build", "-o", filepath.Join(project, "build", "localapp"), "./testdata/localapp")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		t.Fatalf("failed to build helper: %v", err)
	}

	port := freePort(t)
	// Run the helper the way a developer would: a relative path from the project.
	app := exec.CommandContext(t.Context(), "build/localapp", strconv.Itoa(port))
	app.Dir = project
	if err := app.Start(); err != nil {
		t.Fatalf("failed to start local app: %v", err)
	}
	t.Cleanup(func() {
		_ = app.Process.Kill()
		_, _ = app.Process.Wait()
	})
	waitForPort(t, port, 5*time.Second)

	return port
}
