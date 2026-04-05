package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsDevProcess(t *testing.T) {
	if !IsDevProcess("node", "node server.js") {
		t.Fatal("expected node to be treated as a dev process")
	}
	if IsDevProcess("Spotify", "Spotify") {
		t.Fatal("expected Spotify to be filtered out")
	}
}

func TestDetectFrameworkFromCommand(t *testing.T) {
	if got := DetectFrameworkFromCommand("node ./node_modules/.bin/next dev", "node"); got != "Next.js" {
		t.Fatalf("expected Next.js, got %q", got)
	}
	if got := DetectFrameworkFromCommand("uvicorn app:app --reload", "python3"); got != "FastAPI" {
		t.Fatalf("expected FastAPI, got %q", got)
	}
}

func TestDetectFramework(t *testing.T) {
	dir := t.TempDir()
	packageJSON := `{
  "dependencies": {
    "vite": "latest",
    "react": "latest"
  }
}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0o644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	if got := DetectFramework(dir); got != "Vite" {
		t.Fatalf("expected Vite, got %q", got)
	}
}

func TestFindProjectRoot(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("failed to create nested dirs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	if got := FindProjectRoot(nested); got != root {
		t.Fatalf("expected %q, got %q", root, got)
	}
}
