package render

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/LarsEckart/ports/scanner"
)

func TestDisplayPortTableMemory(t *testing.T) {
	var output bytes.Buffer
	DisplayPortTable(&output, []scanner.PortInfo{
		{Port: 3000, PID: 42, ProcessName: "node", MemoryKB: 128 * 1024},
		{Port: 3001, PID: 42, ProcessName: "node", MemoryKB: 128 * 1024},
		{Port: 4000, PID: 43, ProcessName: "python"},
	}, false)

	var rows [][]string
	for _, line := range strings.Split(output.String(), "\n") {
		if strings.Count(line, "│") == 9 {
			cells := strings.Split(line, "│")
			for i := range cells {
				cells[i] = strings.TrimSpace(cells[i])
			}
			rows = append(rows, cells)
		}
	}
	if len(rows) != 4 {
		t.Fatalf("got %d table rows, want header and 3 port rows: %s", len(rows), output.String())
	}
	for i, want := range []string{"MEM", "128.0 MB", "128.0 MB", "—"} {
		if len(rows[i]) != 10 {
			t.Fatalf("row %d has unexpected cells: %v", i, rows[i])
		}
		if got := rows[i][4]; got != want {
			t.Errorf("row %d memory = %q, want %q", i, got, want)
		}
	}
}

func TestRenderTableWidthsAndBorders(t *testing.T) {
	headers := []string{"PORT", "NAME"}
	rows := [][]string{{":1", "node"}, {":3000", "worker"}}
	if got, want := tableWidths(headers, rows), []int{5, 6}; !reflect.DeepEqual(got, want) {
		t.Fatalf("widths = %v, want %v", got, want)
	}
	lines := strings.Split(strings.TrimSuffix(renderTable(headers, rows), "\n"), "\n")
	if len(lines) != 7 {
		t.Fatalf("got %d lines, want 7: %q", len(lines), lines)
	}
	for i, line := range lines {
		if width := lipgloss.Width(line); width != 18 {
			t.Errorf("line %d width = %d, want 18: %q", i, width, line)
		}
	}
	for i, want := range []string{"┌", "│", "├", "│", "├", "│", "└"} {
		if !strings.Contains(lines[i], want) {
			t.Errorf("line %d missing %q: %q", i, want, lines[i])
		}
	}
}
