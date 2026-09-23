package scanner

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	ErrKillTargetAmbiguous = errors.New("kill target is ambiguous")
	batchPSLineRE          = regexp.MustCompile(`^\s*(\d+)\s+(\d+)\s+(\S+)\s+(\d+)\s+(\S+)\s+(.*)$`)
	processLineRE          = regexp.MustCompile(`^\s*(\d+)\s+([\d.]+)\s+(\d+)\s+(\S+)\s+(.*)$`)
	dockerPortRE           = regexp.MustCompile(`:(\d+)->`)
	startTimeLayout        = "Mon Jan _2 15:04:05 2006"
)

func GetListeningPorts(ctx context.Context, detailed bool) ([]PortInfo, error) {
	output, err := run(ctx, 10*time.Second, "lsof", "-nP", "-iTCP", "-sTCP:LISTEN")
	if err != nil {
		return nil, fmt.Errorf("list listening ports: %w", err)
	}

	entries, pids := parseListeningPorts(output)
	if len(entries) == 0 {
		return nil, nil
	}

	metadata, err := loadPortMetadata(ctx, entries, pids)
	if err != nil {
		return nil, err
	}
	for i := range entries {
		entry := &entries[i]
		enrichPort(entry, metadata)
		if err := maybeEnrichPortDetail(ctx, entry, detailed); err != nil {
			return nil, err
		}
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Port < entries[j].Port })
	return entries, nil
}

func loadPortMetadata(ctx context.Context, entries []PortInfo, pids []int) (portMetadata, error) {
	psMap, err := batchPS(ctx, pids)
	if err != nil {
		return portMetadata{}, fmt.Errorf("load process metadata: %w", err)
	}
	parentPSMap, err := batchParentPS(ctx, psMap)
	if err != nil {
		return portMetadata{}, fmt.Errorf("load parent process metadata: %w", err)
	}
	cwdMap, err := batchCWD(ctx, pids)
	if err != nil {
		return portMetadata{}, fmt.Errorf("load process directories: %w", err)
	}
	dockerMap := map[int]dockerInfo{}
	if containsDockerProcess(entries) {
		dockerMap, _ = batchDockerInfo(ctx)
	}
	return portMetadata{psMap, parentPSMap, cwdMap, dockerMap}, nil
}

func maybeEnrichPortDetail(ctx context.Context, entry *PortInfo, detailed bool) error {
	if !detailed {
		return nil
	}
	return enrichPortDetail(ctx, entry)
}

func parseListeningPorts(output string) ([]PortInfo, []int) {
	lines := splitLines(output)
	if len(lines) <= 1 {
		return nil, nil
	}

	entries := make([]PortInfo, 0)
	seenPorts := map[int]struct{}{}
	pids := make([]int, 0)
	seenPIDs := map[int]struct{}{}
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}

		nameField := strings.Join(fields[8:], " ")
		port, ok := parseListenPort(nameField)
		if !ok {
			continue
		}
		if _, ok := seenPorts[port]; ok {
			continue
		}
		seenPorts[port] = struct{}{}

		entries = append(entries, PortInfo{
			Port:        port,
			PID:         pid,
			ProcessName: fields[0],
			RawName:     fields[0],
			Status:      PortStatusHealthy,
		})

		if _, ok := seenPIDs[pid]; !ok {
			seenPIDs[pid] = struct{}{}
			pids = append(pids, pid)
		}
	}

	return entries, pids
}

func containsDockerProcess(entries []PortInfo) bool {
	for _, entry := range entries {
		if isDockerProcess(entry.ProcessName) {
			return true
		}
	}
	return false
}

type portMetadata struct {
	processes map[int]psInfo
	parents   map[int]psInfo
	dirs      map[int]string
	docker    map[int]dockerInfo
}

func enrichPort(entry *PortInfo, metadata portMetadata) {
	if ps, ok := metadata.processes[entry.PID]; ok {
		enrichPortProcess(entry, ps, metadata.parents)
	}
	if docker, ok := metadata.docker[entry.Port]; ok {
		entry.ProjectName = docker.Name
		entry.Framework = DetectFrameworkFromImage(docker.Image)
		entry.ProcessName = "docker"
	}
	if cwd, ok := metadata.dirs[entry.PID]; ok {
		enrichPortDirectory(entry, cwd)
	}
}

func enrichPortProcess(entry *PortInfo, ps psInfo, parents map[int]psInfo) {
	entry.Command = ps.Command
	entry.ProcessName = processNameFromCommand(ps.Command, entry.ProcessName)
	if parent, ok := parents[ps.PPID]; ok {
		entry.ParentCommand = parent.Command
		if label, ok := uvxProjectLabel(parent.Command); ok {
			entry.ProjectName = label
		}
	}
	entry.MemoryKB = ps.RSSKB
	entry.Uptime = elapsedDuration(ps.Elapsed)
	entry.Framework = DetectFrameworkFromCommand(ps.Command, entry.ProcessName)

	switch {
	case strings.Contains(ps.Stat, "Z"):
		entry.Status = PortStatusZombie
	case ps.PPID == 1 && IsDevProcess(entry.ProcessName, ps.Command):
		entry.Status = PortStatusOrphaned
	default:
		entry.Status = PortStatusHealthy
	}
}

func enrichPortDirectory(entry *PortInfo, cwd string) {
	projectRoot := FindProjectRoot(cwd)
	if projectRoot == "/" {
		return
	}
	entry.CWD = projectRoot
	if entry.ProjectName == "" {
		entry.ProjectName = projectLabel(projectRoot)
	}
	if entry.Framework == "" {
		entry.Framework = DetectFramework(projectRoot)
	}
}

// processNameFromCommand prefers ps because lsof shortens COMMAND names on macOS.
func processNameFromCommand(command, fallback string) string {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return fallback
	}
	return filepath.Base(fields[0])
}

func parseListenPort(nameField string) (int, bool) {
	endpoint := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(nameField), "(LISTEN)"))
	_, portText, found := strings.CutLast(endpoint, ":")
	if !found || portText == "" {
		return 0, false
	}

	port, err := strconv.Atoi(portText)
	if err != nil {
		return 0, false
	}
	if port <= 0 || port > 65535 {
		return 0, false
	}
	return port, true
}

func GetPortDetails(ctx context.Context, targetPort int) (*PortInfo, error) {
	ports, err := GetListeningPorts(ctx, true)
	if err != nil {
		return nil, err
	}
	for i := range ports {
		if ports[i].Port == targetPort {
			return &ports[i], nil
		}
	}
	return nil, nil
}

func GetAllProcesses(ctx context.Context) ([]ProcessInfo, error) {
	output, err := run(ctx, 5*time.Second, "ps", "-axo", "pid=,pcpu=,rss=,etime=,command=")
	if err != nil {
		return nil, fmt.Errorf("list processes: %w", err)
	}

	entries, pids := parseProcesses(output)
	cwdMap, err := batchCWD(ctx, pids)
	if err != nil {
		return nil, fmt.Errorf("load process directories: %w", err)
	}

	for i := range entries {
		if cwd, ok := cwdMap[entries[i].PID]; ok {
			enrichProcessDirectory(&entries[i], cwd)
		}
	}
	return entries, nil
}

func parseProcesses(output string) ([]ProcessInfo, []int) {
	entries := make([]ProcessInfo, 0)
	pids := make([]int, 0)
	for _, line := range splitLines(output) {
		match := processLineRE.FindStringSubmatch(line)
		if len(match) != 6 {
			continue
		}
		pid, _ := strconv.Atoi(match[1])
		if pid <= 1 {
			continue
		}
		command := match[5]
		parts := strings.Fields(command)
		if len(parts) == 0 {
			continue
		}
		processName := filepath.Base(parts[0])
		cpu, _ := strconv.ParseFloat(match[2], 64)
		rss, _ := strconv.Atoi(match[3])
		entries = append(entries, ProcessInfo{
			PID:         pid,
			ProcessName: processName,
			Command:     command,
			Description: SummarizeCommand(command, processName),
			CPU:         cpu,
			MemoryKB:    rss,
			Framework:   DetectFrameworkFromCommand(command, processName),
			Uptime:      elapsedDuration(match[4]),
		})
		if !isDockerProcess(processName) {
			pids = append(pids, pid)
		}
	}
	return entries, pids
}

func enrichProcessDirectory(entry *ProcessInfo, cwd string) {
	projectRoot := FindProjectRoot(cwd)
	if projectRoot == "/" {
		return
	}
	entry.CWD = projectRoot
	entry.ProjectName = projectLabel(projectRoot)
	if entry.Framework == "" {
		entry.Framework = DetectFramework(projectRoot)
	}
}

func FindOrphanedProcesses(ctx context.Context) ([]PortInfo, error) {
	ports, err := GetListeningPorts(ctx, false)
	if err != nil {
		return nil, err
	}

	orphaned := make([]PortInfo, 0)
	for _, port := range ports {
		if port.Status == PortStatusOrphaned || port.Status == PortStatusZombie {
			orphaned = append(orphaned, port)
		}
	}
	return orphaned, nil
}

func ResolveKillTarget(ctx context.Context, n int) (*KillTarget, error) {
	if n < 1 {
		return nil, nil
	}

	portTarget, err := ResolveKillPort(ctx, n)
	if err != nil {
		return nil, err
	}
	pidTarget := ResolveKillPID(n)

	switch {
	case portTarget != nil && pidTarget != nil && portTarget.PID != pidTarget.PID:
		return nil, ErrKillTargetAmbiguous
	case portTarget != nil:
		return portTarget, nil
	case pidTarget != nil:
		return pidTarget, nil
	default:
		return nil, nil
	}
}

func ResolveKillPort(ctx context.Context, port int) (*KillTarget, error) {
	if port < 1 || port > 65535 {
		return nil, nil
	}

	info, err := GetPortDetails(ctx, port)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, nil
	}

	return &KillTarget{PID: info.PID, Via: "port", Port: port, Info: info}, nil
}

func ResolveKillPID(pid int) *KillTarget {
	if pid < 1 || !PIDExists(pid) {
		return nil
	}

	return &KillTarget{PID: pid, Via: "pid"}
}

func PIDExists(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}

func KillProcess(pid int, force bool) error {
	signal := syscall.SIGTERM
	if force {
		signal = syscall.SIGKILL
	}
	if err := syscall.Kill(pid, signal); err != nil {
		return fmt.Errorf("signal pid %d: %w", pid, err)
	}
	return nil
}

func WatchPorts(ctx context.Context, interval time.Duration, callback func(eventType string, info PortInfo)) error {
	previous, err := portSnapshot(ctx)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			current, err := portSnapshot(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}
				return err
			}
			emitPortChanges(previous, current, callback)
			previous = current
		}
	}
}

func portSnapshot(ctx context.Context) (map[int]PortInfo, error) {
	ports, err := GetListeningPorts(ctx, false)
	if err != nil {
		return nil, err
	}
	current := make(map[int]PortInfo, len(ports))
	for _, port := range ports {
		current[port.Port] = port
	}
	return current, nil
}

func emitPortChanges(previous, current map[int]PortInfo, callback func(string, PortInfo)) {
	emitMissingPorts(current, previous, "new", callback)
	emitMissingPorts(previous, current, "removed", callback)
}

func emitMissingPorts(source, other map[int]PortInfo, kind string, callback func(string, PortInfo)) {
	for port, info := range source {
		if _, ok := other[port]; !ok {
			callback(kind, info)
		}
	}
}

func CollapseDockerProcesses(processes []ProcessInfo) []ProcessInfo {
	docker := make([]ProcessInfo, 0)
	nonDocker := make([]ProcessInfo, 0, len(processes))
	for _, process := range processes {
		if isDockerProcess(process.ProcessName) {
			docker = append(docker, process)
			continue
		}
		nonDocker = append(nonDocker, process)
	}

	if len(docker) == 0 {
		return nonDocker
	}

	var cpu float64
	var memoryKB int
	for _, process := range docker {
		cpu += process.CPU
		memoryKB += process.MemoryKB
	}

	nonDocker = append(nonDocker, ProcessInfo{
		PID:         docker[0].PID,
		ProcessName: "Docker",
		Description: fmt.Sprintf("%d processes", len(docker)),
		CPU:         cpu,
		MemoryKB:    memoryKB,
		Framework:   "Docker",
		Uptime:      docker[0].Uptime,
	})

	return nonDocker
}

func batchParentPS(ctx context.Context, processes map[int]psInfo) (map[int]psInfo, error) {
	parentPIDs := make([]int, 0, len(processes))
	seen := map[int]struct{}{}
	for _, process := range processes {
		if process.PPID <= 1 {
			continue
		}
		if _, ok := seen[process.PPID]; ok {
			continue
		}
		seen[process.PPID] = struct{}{}
		parentPIDs = append(parentPIDs, process.PPID)
	}

	return batchPS(ctx, parentPIDs)
}

func batchPS(ctx context.Context, pids []int) (map[int]psInfo, error) {
	result := map[int]psInfo{}
	if len(pids) == 0 {
		return result, nil
	}

	args := []string{"-p", joinInts(pids), "-o", "pid=,ppid=,stat=,rss=,etime=,command="}
	output, err := run(ctx, 5*time.Second, "ps", args...)
	if err != nil {
		return nil, err
	}

	for _, line := range splitLines(output) {
		match := batchPSLineRE.FindStringSubmatch(line)
		if len(match) != 7 {
			continue
		}
		pid, _ := strconv.Atoi(match[1])
		ppid, _ := strconv.Atoi(match[2])
		rss, _ := strconv.Atoi(match[4])
		result[pid] = psInfo{
			PPID:    ppid,
			Stat:    match[3],
			RSSKB:   rss,
			Elapsed: match[5],
			Command: match[6],
		}
	}

	return result, nil
}

func batchCWD(ctx context.Context, pids []int) (map[int]string, error) {
	result := map[int]string{}
	if len(pids) == 0 {
		return result, nil
	}

	args := []string{"-a", "-d", "cwd", "-p", joinInts(pids)}
	output, err := run(ctx, 10*time.Second, "lsof", args...)
	if err != nil && strings.TrimSpace(output) == "" {
		return result, nil
	}

	result = parseCWDOutput(output)
	return result, nil
}

func parseCWDOutput(output string) map[int]string {
	result := map[int]string{}
	lines := splitLines(output)
	if len(lines) <= 1 {
		return result
	}

	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}
		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		path := strings.Join(fields[8:], " ")
		if strings.HasPrefix(path, "/") {
			result[pid] = path
		}
	}

	return result
}

func batchDockerInfo(ctx context.Context) (map[int]dockerInfo, error) {
	output, err := run(ctx, 5*time.Second, "docker", "ps", "--format", "{{.Ports}}\t{{.Names}}\t{{.Image}}")
	if err != nil {
		return map[int]dockerInfo{}, nil
	}
	return parseDockerInfo(output), nil
}

func parseDockerInfo(output string) map[int]dockerInfo {
	result := map[int]dockerInfo{}
	for _, line := range splitLines(output) {
		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			continue
		}
		addDockerPorts(result, parts[0], dockerInfo{Name: parts[1], Image: parts[2]})
	}
	return result
}

func addDockerPorts(result map[int]dockerInfo, ports string, info dockerInfo) {
	for _, match := range dockerPortRE.FindAllStringSubmatch(ports, -1) {
		if len(match) != 2 {
			continue
		}
		if port, err := strconv.Atoi(match[1]); err == nil {
			result[port] = info
		}
	}
}

func enrichPortDetail(ctx context.Context, port *PortInfo) error {
	startTime, err := getStartTime(ctx, port.PID)
	if err != nil {
		return err
	}
	if startTime != nil {
		port.StartTime = startTime
	}

	if port.CWD != "" {
		branch, err := currentGitBranch(ctx, port.CWD)
		if err == nil {
			port.GitBranch = branch
		}
	}

	tree, err := getProcessTree(ctx, port.PID)
	if err != nil {
		return err
	}
	port.ProcessTree = tree
	return nil
}

func getStartTime(ctx context.Context, pid int) (*time.Time, error) {
	output, err := run(ctx, 3*time.Second, "ps", "-p", strconv.Itoa(pid), "-o", "lstart=")
	if err != nil {
		return nil, nil
	}
	text := strings.TrimSpace(output)
	if text == "" {
		return nil, nil
	}
	t, err := time.Parse(startTimeLayout, text)
	if err != nil {
		return nil, nil
	}
	return &t, nil
}

func currentGitBranch(ctx context.Context, dir string) (string, error) {
	output, err := run(ctx, 3*time.Second, "git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func getProcessTree(ctx context.Context, pid int) ([]ProcessNode, error) {
	output, err := run(ctx, 5*time.Second, "ps", "-axo", "pid=,ppid=,comm=")
	if err != nil {
		return nil, err
	}

	return processTree(parseProcessNodes(output), pid), nil
}

func parseProcessNodes(output string) map[int]ProcessNode {
	all := map[int]ProcessNode{}
	for _, line := range splitLines(output) {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		childPID, err1 := strconv.Atoi(fields[0])
		ppid, err2 := strconv.Atoi(fields[1])
		if err1 != nil || err2 != nil {
			continue
		}
		all[childPID] = ProcessNode{
			PID:  childPID,
			PPID: ppid,
			Name: filepath.Base(strings.Join(fields[2:], " ")),
		}
	}
	return all
}

func processTree(all map[int]ProcessNode, pid int) []ProcessNode {
	tree := make([]ProcessNode, 0, 8)
	current := pid
	for depth := 0; current > 1 && depth < 8; depth++ {
		node, ok := all[current]
		if !ok {
			break
		}
		tree = append(tree, node)
		current = node.PPID
	}
	return tree
}

func run(ctx context.Context, timeout time.Duration, name string, args ...string) (string, error) {
	cmdCtx := ctx
	var cancel context.CancelFunc
	if timeout > 0 {
		cmdCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(cmdCtx, name, args...)
	output, err := cmd.Output()
	if err == nil {
		return string(output), nil
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr := strings.TrimSpace(string(exitErr.Stderr))
		if stderr == "" {
			return string(output), err
		}
		return string(output), fmt.Errorf("%s: %s", name, stderr)
	}

	return string(output), err
}

func joinInts(values []int) string {
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = strconv.Itoa(value)
	}
	return strings.Join(parts, ",")
}

func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func isDockerProcess(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasPrefix(lower, "com.docke") || strings.HasPrefix(lower, "docker")
}

func elapsedDuration(value string) time.Duration {
	if value == "" {
		return 0
	}

	days := 0
	clock := value
	if strings.Contains(value, "-") {
		parts := strings.SplitN(value, "-", 2)
		days, _ = strconv.Atoi(parts[0])
		clock = parts[1]
	}

	segments := strings.Split(clock, ":")
	var hours, minutes, seconds int
	switch len(segments) {
	case 2:
		minutes, _ = strconv.Atoi(segments[0])
		seconds, _ = strconv.Atoi(segments[1])
	case 3:
		hours, _ = strconv.Atoi(segments[0])
		minutes, _ = strconv.Atoi(segments[1])
		seconds, _ = strconv.Atoi(segments[2])
	default:
		return 0
	}

	return time.Duration(days)*24*time.Hour + time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second
}
