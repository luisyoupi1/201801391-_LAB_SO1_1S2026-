package model

import "time"

type Memory struct {
	TotalKB uint64 `json:"total_kb"`
	FreeKB  uint64 `json:"free_kb"`
	UsedKB  uint64 `json:"used_kb"`
}

type Process struct {
	PID           int     `json:"pid"`
	Name          string  `json:"name"`
	CommandLine   string  `json:"cmdline"`
	VSZKB         uint64  `json:"vsz_kb"`
	RSSKB         uint64  `json:"rss_kb"`
	MemoryPercent float64 `json:"memory_percent"`
	CPUPercent    float64 `json:"cpu_percent"`
}

type Snapshot struct {
	Carnet    string    `json:"carnet"`
	Memory    Memory    `json:"memory"`
	Processes []Process `json:"processes"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

type Container struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	PID     int               `json:"pid"`
	Running bool              `json:"running"`
	Labels  map[string]string `json:"labels"`
	Profile string            `json:"profile"`
}

type ContainerMetric struct {
	Container     Container `json:"container"`
	VSZKB         uint64    `json:"vsz_kb"`
	RSSKB         uint64    `json:"rss_kb"`
	MemoryPercent float64   `json:"memory_percent"`
	CPUPercent    float64   `json:"cpu_percent"`
	Score         float64   `json:"score"`
}

type KillEvent struct {
	Origin      string    `json:"origin,omitempty"`
	SourcePID   uint32    `json:"source_pid"`
	TargetPID   uint32    `json:"target_pid"`
	Signal      int32     `json:"signal"`
	TimestampNS uint64    `json:"timestamp_ns"`
	Command     string    `json:"command"`
	ObservedAt  time.Time `json:"observed_at"`
}

type Deletion struct {
	Container ContainerMetric `json:"container"`
	Event     KillEvent       `json:"event"`
	DeletedAt time.Time       `json:"deleted_at"`
}
