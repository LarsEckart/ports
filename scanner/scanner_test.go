package scanner

import (
	"context"
	"net"
	"path/filepath"
	"testing"
	"time"
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

func TestParseListenPort(t *testing.T) {
	tests := []struct {
		name      string
		nameField string
		want      int
		wantOK    bool
	}{
		{
			name:      "wildcard IPv4",
			nameField: "TCP *:60543 (LISTEN)",
			want:      60543,
			wantOK:    true,
		},
		{
			name:      "loopback IPv4",
			nameField: "TCP 127.0.0.1:8080 (LISTEN)",
			want:      8080,
			wantOK:    true,
		},
		{
			name:      "loopback IPv6",
			nameField: "TCP [::1]:42050 (LISTEN)",
			want:      42050,
			wantOK:    true,
		},
		{
			name:      "scoped IPv6",
			nameField: "TCP [fe80::1%lo0]:12345 (LISTEN)",
			want:      12345,
			wantOK:    true,
		},
		{
			name:      "missing port",
			nameField: "TCP [::1] (LISTEN)",
			wantOK:    false,
		},
		{
			name:      "non-numeric port",
			nameField: "TCP 127.0.0.1:http (LISTEN)",
			wantOK:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseListenPort(tt.nameField)
			if ok != tt.wantOK {
				t.Fatalf("expected ok=%v, got %v", tt.wantOK, ok)
			}
			if got != tt.want {
				t.Fatalf("expected port %d, got %d", tt.want, got)
			}
		})
	}
}

func TestWatchPortsSkipsInitialListeners(t *testing.T) {
	initialListener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start initial listener: %v", err)
	}
	defer func() { _ = initialListener.Close() }()

	initialPort := initialListener.Addr().(*net.TCPAddr).Port
	var newPort int
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	type event struct {
		kind string
		port int
	}
	events := make(chan event, 4)
	errCh := make(chan error, 1)

	go func() {
		errCh <- WatchPorts(ctx, 100*time.Millisecond, func(eventType string, info PortInfo) {
			if info.Port == initialPort || info.Port == newPort {
				events <- event{kind: eventType, port: info.Port}
			}
		})
	}()

	select {
	case event := <-events:
		t.Fatalf("did not expect an initial event for existing listener: %+v", event)
	case <-time.After(350 * time.Millisecond):
	}

	newListener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start new listener: %v", err)
	}
	defer func() { _ = newListener.Close() }()

	newPort = newListener.Addr().(*net.TCPAddr).Port

	deadline := time.After(3 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for new listener on port %d", newPort)
		case event := <-events:
			if event.port != newPort {
				continue
			}
			if event.kind != "new" {
				t.Fatalf("expected new event for port %d, got %+v", newPort, event)
			}
			cancel()
			if err := <-errCh; err != nil {
				t.Fatalf("watch returned error: %v", err)
			}
			return
		}
	}
}
