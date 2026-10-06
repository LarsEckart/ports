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

func TestFrameworkDetectionKeepsPriority(t *testing.T) {
	if got := DetectFrameworkFromImage("redis-postgres"); got != "PostgreSQL" {
		t.Fatalf("image priority: got %q", got)
	}
	if got := DetectFrameworkFromCommand("vite next", "node"); got != "Next.js" {
		t.Fatalf("command priority: got %q", got)
	}

	dir := t.TempDir()
	packageJSON := `{"dependencies":{"react":"latest","next":"latest"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := DetectFramework(dir); got != "Next.js" {
		t.Fatalf("dependency priority: got %q", got)
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

func TestProjectLabel(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{path: "/Users/lars/GitHub/ports/frontend", want: "ports/frontend"},
		{path: "/frontend", want: "frontend"},
	}

	for _, tt := range tests {
		if got := projectLabel(tt.path); got != tt.want {
			t.Errorf("projectLabel(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestPortBuiltInProject(t *testing.T) {
	root := "/Users/lars/GitHub/SmellCheck"
	tests := []struct {
		name string
		port PortInfo
		want bool
	}{
		{
			name: "relative build path inside the project",
			port: PortInfo{CWD: root, Command: "build/smellcheck --port 8000"},
			want: true,
		},
		{
			name: "absolute build path inside the project",
			port: PortInfo{CWD: root, Command: root + "/build/smellcheck"},
			want: true,
		},
		{
			name: "installed binary outside the project",
			port: PortInfo{CWD: root, Command: "/opt/homebrew/bin/node server.js"},
			want: false,
		},
		{
			name: "bare name goes through PATH lookup",
			port: PortInfo{CWD: root, Command: "smellcheck"},
			want: false,
		},
		{
			name: "sibling directory outside the project",
			port: PortInfo{CWD: root, Command: "../other/build/app"},
			want: false,
		},
		{
			name: "unknown project directory",
			port: PortInfo{Command: "build/smellcheck"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.port.BuiltInProject(); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}

	process := ProcessInfo{CWD: root, Command: "build/smellcheck"}
	if !process.BuiltInProject() {
		t.Fatal("expected a process entry to count as built in project")
	}
}

func TestFindProjectRootStopsAtHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	// A stray manifest in home must not claim every process as a project.
	if err := os.WriteFile(filepath.Join(home, "package.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}
	work := filepath.Join(home, "scratch", "app")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatalf("failed to create work dir: %v", err)
	}

	if got := FindProjectRoot(work); got != work {
		t.Fatalf("expected %q, got %q", work, got)
	}
}

func TestFindProjectRootDetectsRepoAndRocApp(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create .git: %v", err)
	}
	nested := filepath.Join(repo, "internal", "server")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("failed to create nested dirs: %v", err)
	}
	if got := FindProjectRoot(nested); got != repo {
		t.Fatalf("repo root: expected %q, got %q", repo, got)
	}

	rocApp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(rocApp, "app"), 0o755); err != nil {
		t.Fatalf("failed to create app dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rocApp, "app", "main.roc"), []byte("app \"demo\"\n"), 0o644); err != nil {
		t.Fatalf("failed to write app/main.roc: %v", err)
	}
	if got := FindProjectRoot(filepath.Join(rocApp, "build")); got != rocApp {
		t.Fatalf("roc root: expected %q, got %q", rocApp, got)
	}
}

func TestDetectFrameworkDetectsRoc(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "app"), 0o755); err != nil {
		t.Fatalf("failed to create app dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "app", "main.roc"), []byte("app \"demo\"\n"), 0o644); err != nil {
		t.Fatalf("failed to write app/main.roc: %v", err)
	}

	if got := DetectFramework(dir); got != "Roc" {
		t.Fatalf("expected Roc, got %q", got)
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
