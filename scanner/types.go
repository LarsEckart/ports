package scanner

import "time"

type PortStatus uint8

const (
	PortStatusUnknown PortStatus = iota
	PortStatusHealthy
	PortStatusOrphaned
	PortStatusZombie
)

type PortInfo struct {
	Port        int
	PID         int
	ProcessName string
	RawName     string
	Command     string
	CWD         string
	ProjectName string
	Framework   string
	Uptime      time.Duration
	Status      PortStatus
	MemoryKB    int
	GitBranch   string
	StartTime   *time.Time
	ProcessTree []ProcessNode
}

type ProcessNode struct {
	PID  int
	PPID int
	Name string
}

type ProcessInfo struct {
	PID         int
	ProcessName string
	Command     string
	Description string
	CPU         float64
	MemoryKB    int
	CWD         string
	ProjectName string
	Framework   string
	Uptime      time.Duration
}

type KillTarget struct {
	PID  int
	Via  string
	Port int
	Info *PortInfo
}

type dockerInfo struct {
	Name  string
	Image string
}

type psInfo struct {
	PPID    int
	Stat    string
	RSSKB   int
	Elapsed string
	Command string
}
