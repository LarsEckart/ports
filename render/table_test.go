package render

import (
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

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
