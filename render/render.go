package render

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/LarsEckart/ports/scanner"
	"github.com/charmbracelet/lipgloss"
)

var (
	borderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	mutedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	whiteStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	blueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	yellowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("221"))
	greenStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	redStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	cyanStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("51"))
	pinkStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("213"))
)

var frameworkStyles = map[string]lipgloss.Style{
	"Next.js":       whiteStyle,
	"Vite":          yellowStyle,
	"React":         cyanStyle,
	"Vue":           greenStyle,
	"Angular":       redStyle,
	"Svelte":        lipgloss.NewStyle().Foreground(lipgloss.Color("202")),
	"SvelteKit":     lipgloss.NewStyle().Foreground(lipgloss.Color("202")),
	"Express":       mutedStyle,
	"Fastify":       whiteStyle,
	"NestJS":        redStyle,
	"Nuxt":          greenStyle,
	"Remix":         blueStyle,
	"Astro":         pinkStyle,
	"Django":        greenStyle,
	"Flask":         whiteStyle,
	"FastAPI":       cyanStyle,
	"Rails":         redStyle,
	"Gatsby":        pinkStyle,
	"Go":            cyanStyle,
	"Rust":          lipgloss.NewStyle().Foreground(lipgloss.Color("215")),
	"Ruby":          redStyle,
	"Python":        yellowStyle,
	"Node.js":       greenStyle,
	"Java":          redStyle,
	"Hono":          lipgloss.NewStyle().Foreground(lipgloss.Color("208")),
	"Koa":           whiteStyle,
	"Webpack":       blueStyle,
	"esbuild":       yellowStyle,
	"Parcel":        lipgloss.NewStyle().Foreground(lipgloss.Color("179")),
	"Docker":        blueStyle,
	"PostgreSQL":    blueStyle,
	"Redis":         redStyle,
	"MySQL":         blueStyle,
	"MongoDB":       greenStyle,
	"nginx":         greenStyle,
	"LocalStack":    whiteStyle,
	"RabbitMQ":      lipgloss.NewStyle().Foreground(lipgloss.Color("208")),
	"Kafka":         whiteStyle,
	"Elasticsearch": yellowStyle,
	"MinIO":         redStyle,
}

func DisplayPortTable(w io.Writer, ports []scanner.PortInfo, filtered bool) {
	renderHeader(w)
	if len(ports) == 0 {
		fmt.Fprintln(w, mutedStyle.Render("  No active listening ports found."))
		fmt.Fprintln(w)
		if filtered {
			fmt.Fprintln(w, mutedStyle.Render("  Run ports --all to show everything."))
			fmt.Fprintln(w)
		}
		return
	}

	rows := make([][]string, 0, len(ports))
	for _, port := range ports {
		rows = append(rows, []string{
			whiteStyle.Bold(true).Render(fmt.Sprintf(":%d", port.Port)),
			whiteStyle.Render(orDash(truncate(port.ProcessName, 15))),
			mutedStyle.Render(fmt.Sprintf("%d", port.PID)),
			styledProject(port.ProjectName),
			formatFramework(port.Framework),
			styledUptime(port.Uptime),
			formatStatus(port.Status),
		})
	}

	headers := []string{"PORT", "PROCESS", "PID", "PROJECT", "FRAMEWORK", "UPTIME", "STATUS"}
	fmt.Fprintln(w, renderTable(headers, rows))
	fmt.Fprintln(w)

	hint := mutedStyle.Render(fmt.Sprintf("  %d port%s active", len(ports), plural(len(ports)))) +
		mutedStyle.Render("  ·  Run ") + cyanStyle.Render("ports <number>") + mutedStyle.Render(" for details")
	if filtered {
		hint += mutedStyle.Render("  ·  ") + cyanStyle.Render("--all") + mutedStyle.Render(" to show everything")
	}
	fmt.Fprintln(w, hint)
	fmt.Fprintln(w)
}

func DisplayProcessTable(w io.Writer, processes []scanner.ProcessInfo, filtered bool) {
	renderHeader(w)
	if len(processes) == 0 {
		fmt.Fprintln(w, mutedStyle.Render("  No matching processes found."))
		fmt.Fprintln(w)
		return
	}

	rows := make([][]string, 0, len(processes))
	for _, process := range processes {
		rows = append(rows, []string{
			mutedStyle.Render(fmt.Sprintf("%d", process.PID)),
			whiteStyle.Bold(true).Render(truncate(process.ProcessName, 15)),
			formatCPU(process.CPU),
			styledMemory(process.MemoryKB),
			styledProject(process.ProjectName),
			formatFramework(process.Framework),
			styledUptime(process.Uptime),
			mutedStyle.Render(truncate(orDash(process.Description), 32)),
		})
	}

	headers := []string{"PID", "PROCESS", "CPU%", "MEM", "PROJECT", "FRAMEWORK", "UPTIME", "WHAT"}
	fmt.Fprintln(w, renderTable(headers, rows))
	fmt.Fprintln(w)

	hint := mutedStyle.Render(fmt.Sprintf("  %d process%s", len(processes), plural(len(processes))))
	if filtered {
		hint += mutedStyle.Render("  ·  ") + cyanStyle.Render("--all") + mutedStyle.Render(" to show everything")
	}
	fmt.Fprintln(w, hint)
	fmt.Fprintln(w)
}

func DisplayPortDetail(w io.Writer, info *scanner.PortInfo) {
	renderHeader(w)
	if info == nil {
		fmt.Fprintln(w, redStyle.Render("  No process found on that port."))
		fmt.Fprintln(w)
		return
	}

	fmt.Fprintln(w, whiteStyle.Bold(true).Render(fmt.Sprintf("  Port :%d", info.Port)))
	fmt.Fprintln(w, mutedStyle.Render("  ──────────────────────"))
	fmt.Fprintln(w)

	printField(w, "Process", whiteStyle.Bold(true).Render(orDash(info.ProcessName)))
	printField(w, "PID", mutedStyle.Render(fmt.Sprintf("%d", info.PID)))
	printField(w, "Status", formatStatus(info.Status))
	printField(w, "Framework", formatFramework(info.Framework))
	printField(w, "Memory", styledMemory(info.MemoryKB))
	printField(w, "Uptime", styledUptime(info.Uptime))
	if info.StartTime != nil {
		printField(w, "Started", mutedStyle.Render(info.StartTime.In(time.Local).Format(time.RFC1123)))
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, titleStyle.Render("  Location"))
	fmt.Fprintln(w, mutedStyle.Render("  ──────────────────────"))
	printField(w, "Directory", styledPath(info.CWD))
	printField(w, "Project", whiteStyle.Render(orDash(info.ProjectName)))
	printField(w, "Git Branch", styledBranch(info.GitBranch))

	if len(info.ProcessTree) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, titleStyle.Render("  Process Tree"))
		fmt.Fprintln(w, mutedStyle.Render("  ──────────────────────"))
		for i, node := range info.ProcessTree {
			indent := strings.Repeat("  ", i)
			prefix := "└─"
			if i == 0 {
				prefix = "→"
			}
			name := mutedStyle.Render(node.Name)
			if node.PID == info.PID {
				name = whiteStyle.Bold(true).Render(node.Name)
			}
			fmt.Fprintf(w, "  %s%s %s %s\n", indent, mutedStyle.Render(prefix), name, mutedStyle.Render(fmt.Sprintf("(%d)", node.PID)))
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, mutedStyle.Render("  Kill this process with ")+cyanStyle.Render(fmt.Sprintf("ports kill %d", info.Port)))
	fmt.Fprintln(w)
}

func DisplayCleanResults(w io.Writer, orphaned []scanner.PortInfo, killed, failed []int) {
	renderHeader(w)
	if len(orphaned) == 0 {
		fmt.Fprintln(w, greenStyle.Render("  ✓ No orphaned or zombie dev processes found."))
		fmt.Fprintln(w)
		return
	}

	fmt.Fprintf(w, "%s\n\n", yellowStyle.Bold(true).Render(fmt.Sprintf("  Found %d orphaned/zombie process%s:", len(orphaned), plural(len(orphaned)))))
	for _, port := range orphaned {
		icon := yellowStyle.Render("?")
		if containsInt(killed, port.PID) {
			icon = greenStyle.Render("✓")
		}
		if containsInt(failed, port.PID) {
			icon = redStyle.Render("✕")
		}
		fmt.Fprintf(w, "  %s %s %s %s\n", icon, whiteStyle.Bold(true).Render(fmt.Sprintf(":%d", port.Port)), mutedStyle.Render("—"), mutedStyle.Render(fmt.Sprintf("%s (PID %d)", port.ProcessName, port.PID)))
	}
	fmt.Fprintln(w)
	if len(killed) > 0 {
		fmt.Fprintln(w, greenStyle.Render(fmt.Sprintf("  Cleaned %d process%s.", len(killed), plural(len(killed)))))
	}
	if len(failed) > 0 {
		fmt.Fprintln(w, redStyle.Render(fmt.Sprintf("  Failed to clean %d process%s.", len(failed), plural(len(failed)))))
	}
	fmt.Fprintln(w)
}

func DisplayWatchHeader(w io.Writer) {
	renderHeader(w)
	fmt.Fprintln(w, titleStyle.Render("  Watching for port changes..."))
	fmt.Fprintln(w, mutedStyle.Render("  Press Ctrl+C to stop"))
	fmt.Fprintln(w)
}

func DisplayWatchEvent(w io.Writer, eventType string, info scanner.PortInfo) {
	timestamp := mutedStyle.Render(time.Now().Format("15:04:05"))
	switch eventType {
	case "new":
		project := ""
		if info.ProjectName != "" {
			project = blueStyle.Render(" [" + info.ProjectName + "]")
		}
		framework := ""
		if info.Framework != "" {
			framework = " " + formatFramework(info.Framework)
		}
		fmt.Fprintf(w, "  %s %s %s ← %s%s%s\n", timestamp, greenStyle.Render("▲ NEW"), whiteStyle.Bold(true).Render(fmt.Sprintf(":%d", info.Port)), whiteStyle.Render(info.ProcessName), project, framework)
	case "removed":
		fmt.Fprintf(w, "  %s %s %s\n", timestamp, redStyle.Render("▼ CLOSED"), whiteStyle.Bold(true).Render(fmt.Sprintf(":%d", info.Port)))
	}
}

func renderHeader(w io.Writer) {
	border := strings.Repeat("─", 33)
	fmt.Fprintln(w)
	fmt.Fprintln(w, cyanStyle.Bold(true).Render(" ┌"+border+"┐"))
	fmt.Fprintln(w, cyanStyle.Bold(true).Render(" │")+whiteStyle.Bold(true).Render(padToWidth("  ports", 33))+cyanStyle.Bold(true).Render("│"))
	fmt.Fprintln(w, cyanStyle.Bold(true).Render(" │")+mutedStyle.Render(padToWidth("  what is listening right now", 33))+cyanStyle.Bold(true).Render("│"))
	fmt.Fprintln(w, cyanStyle.Bold(true).Render(" └"+border+"┘"))
	fmt.Fprintln(w)
}

func renderTable(headers []string, rows [][]string) string {
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = max(widths[i], lipgloss.Width(header))
	}
	for _, row := range rows {
		for i, cell := range row {
			widths[i] = max(widths[i], lipgloss.Width(cell))
		}
	}

	var b strings.Builder
	writeBorder(&b, widths, "┌", "┬", "┐")
	writeRow(&b, widths, styleHeaders(headers))
	writeBorder(&b, widths, "├", "┼", "┤")
	for i, row := range rows {
		writeRow(&b, widths, row)
		if i < len(rows)-1 {
			writeBorder(&b, widths, "├", "┼", "┤")
		}
	}
	writeBorder(&b, widths, "└", "┴", "┘")
	return b.String()
}

func writeBorder(b *strings.Builder, widths []int, left, mid, right string) {
	b.WriteString(borderStyle.Render(left))
	for i, width := range widths {
		b.WriteString(borderStyle.Render(strings.Repeat("─", width+2)))
		if i < len(widths)-1 {
			b.WriteString(borderStyle.Render(mid))
		}
	}
	b.WriteString(borderStyle.Render(right))
	b.WriteByte('\n')
}

func writeRow(b *strings.Builder, widths []int, cells []string) {
	b.WriteString(borderStyle.Render("│"))
	for i, cell := range cells {
		b.WriteByte(' ')
		b.WriteString(cell)
		b.WriteString(strings.Repeat(" ", widths[i]-lipgloss.Width(cell)+1))
		b.WriteString(borderStyle.Render("│"))
	}
	b.WriteByte('\n')
}

func styleHeaders(headers []string) []string {
	styled := make([]string, len(headers))
	for i, header := range headers {
		styled[i] = cyanStyle.Bold(true).Render(header)
	}
	return styled
}

func formatFramework(value string) string {
	if value == "" {
		return mutedStyle.Render("—")
	}
	style, ok := frameworkStyles[value]
	if !ok {
		style = whiteStyle
	}
	return style.Render(value)
}

func formatStatus(value scanner.PortStatus) string {
	switch value {
	case scanner.PortStatusHealthy:
		return greenStyle.Render("● healthy")
	case scanner.PortStatusOrphaned:
		return yellowStyle.Render("● orphaned")
	case scanner.PortStatusZombie:
		return redStyle.Render("● zombie")
	default:
		return mutedStyle.Render("● unknown")
	}
}

func formatCPU(value float64) string {
	text := fmt.Sprintf("%.1f", value)
	switch {
	case value > 25:
		return redStyle.Render(text)
	case value > 5:
		return yellowStyle.Render(text)
	default:
		return greenStyle.Render(text)
	}
}

func styledProject(value string) string {
	if value == "" {
		return mutedStyle.Render("—")
	}
	return blueStyle.Render(truncate(value, 22))
}

func styledPath(value string) string {
	if value == "" {
		return mutedStyle.Render("—")
	}
	return blueStyle.Render(value)
}

func styledBranch(value string) string {
	if value == "" {
		return mutedStyle.Render("—")
	}
	return pinkStyle.Render(value)
}

func styledMemory(valueKB int) string {
	if valueKB <= 0 {
		return mutedStyle.Render("—")
	}
	return greenStyle.Render(formatMemory(valueKB))
}

func styledUptime(value time.Duration) string {
	if value <= 0 {
		return mutedStyle.Render("—")
	}
	return yellowStyle.Render(humanDuration(value))
}

func formatMemory(valueKB int) string {
	switch {
	case valueKB > 1024*1024:
		return fmt.Sprintf("%.1f GB", float64(valueKB)/(1024*1024))
	case valueKB > 1024:
		return fmt.Sprintf("%.1f MB", float64(valueKB)/1024)
	default:
		return fmt.Sprintf("%d KB", valueKB)
	}
}

func humanDuration(value time.Duration) string {
	seconds := int(value.Seconds())
	minutes := seconds / 60
	hours := minutes / 60
	days := hours / 24

	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh", days, hours%24)
	case hours > 0:
		return fmt.Sprintf("%dh %dm", hours, minutes%60)
	case minutes > 0:
		return fmt.Sprintf("%dm %ds", minutes, seconds%60)
	default:
		return fmt.Sprintf("%ds", seconds)
	}
}

func printField(w io.Writer, label, value string) {
	fmt.Fprintf(w, "  %-14s %s\n", mutedStyle.Render(label), value)
}

func padToWidth(text string, width int) string {
	return text + strings.Repeat(" ", max(0, width-lipgloss.Width(text)))
}

func truncate(value string, maxWidth int) string {
	if lipgloss.Width(value) <= maxWidth {
		return value
	}
	runes := []rune(value)
	if maxWidth <= 1 {
		return string(runes[:maxWidth])
	}
	return string(runes[:maxWidth-1]) + "…"
}

func plural(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func orDash(value string) string {
	if value == "" {
		return "—"
	}
	return value
}

func containsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
