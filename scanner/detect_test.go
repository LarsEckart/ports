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

func TestIsDevPortDetectsUVXLaunches(t *testing.T) {
	port := PortInfo{
		ProcessName:   "agentsvie",
		Command:       "/Users/lars/.cache/uv/archive-v0/example/lib/python3.13/site-packages/agentsview/bin/agentsview serve",
		ParentCommand: "/Users/lars/.local/bin/uv tool uvx agentsview serve",
	}

	if !IsDevPort(port) {
		t.Fatal("expected a uvx-launched listener to be treated as a dev port")
	}
}

func TestIsDevPortDoesNotPromoteUnknownConsoleScripts(t *testing.T) {
	port := PortInfo{
		ProcessName: "agentsvie",
		Command:     "/Users/lars/.cache/uv/archive-v0/example/lib/python3.13/site-packages/agentsview/bin/agentsview serve",
	}

	if IsDevPort(port) {
		t.Fatal("expected an unknown console script without uvx parent context to stay filtered")
	}
}

func TestIsUVXCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    bool
	}{
		{
			name:    "uv tool uvx shim",
			command: "/Users/lars/.local/bin/uv tool uvx agentsview serve",
			want:    true,
		},
		{
			name:    "direct uvx executable",
			command: "/Users/lars/.local/bin/uvx agentsview serve",
			want:    true,
		},
		{
			name:    "uv command that is not uvx",
			command: "/Users/lars/.local/bin/uv run python app.py",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUVXCommand(tt.command); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestUVXProjectLabel(t *testing.T) {
	got, ok := uvxProjectLabel("/Users/lars/.local/bin/uv tool uvx agentsview serve")
	if !ok {
		t.Fatal("expected uvx project label")
	}
	if got != "uvx agentsview" {
		t.Fatalf("expected project label %q, got %q", "uvx agentsview", got)
	}
}

func TestUVXPackageName(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    string
		wantOK  bool
	}{
		{
			name:    "uv tool uvx shim",
			command: "/Users/lars/.local/bin/uv tool uvx agentsview serve",
			want:    "agentsview",
			wantOK:  true,
		},
		{
			name:    "direct uvx executable",
			command: "/Users/lars/.local/bin/uvx agentsview serve",
			want:    "agentsview",
			wantOK:  true,
		},
		{
			name:    "skips uvx options with values",
			command: "uvx --python 3.13 --isolated agentsview serve",
			want:    "agentsview",
			wantOK:  true,
		},
		{
			name:    "prefers package from --from",
			command: "uvx --from agentsview agentsview serve",
			want:    "agentsview",
			wantOK:  true,
		},
		{
			name:    "cleans version constraints",
			command: "uvx 'agentsview>=1.2' serve",
			want:    "agentsview",
			wantOK:  true,
		},
		{
			name:    "uv command that is not uvx",
			command: "/Users/lars/.local/bin/uv run python app.py",
			wantOK:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := uvxPackageName(tt.command)
			if ok != tt.wantOK {
				t.Fatalf("expected ok=%v, got %v", tt.wantOK, ok)
			}
			if got != tt.want {
				t.Fatalf("expected package %q, got %q", tt.want, got)
			}
		})
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
