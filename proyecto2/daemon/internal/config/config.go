package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Config struct {
	Carnet               string
	ProcPath             string
	LoopInterval         time.Duration
	MetricsAddr          string
	ValkeyAddr           string
	ValkeyPassword       string
	ComposeFile          string
	ProjectRoot          string
	BPFObject            string
	MinLowContainers     int
	MinHighContainers    int
	DeleteConfirmTimeout time.Duration
	EBPFRequired         bool
}

func Load() (Config, error) {
	root := getenv("PROJECT_ROOT", ".")
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return Config{}, fmt.Errorf("resolve project root: %w", err)
	}
	carnet := getenv("CARNET", "201801391")
	interval, err := durationEnv("LOOP_INTERVAL", 30*time.Second)
	if err != nil {
		return Config{}, err
	}
	timeout, err := durationEnv("DELETE_CONFIRM_TIMEOUT", 15*time.Second)
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		Carnet:               carnet,
		ProcPath:             getenv("PROC_PATH", "/proc/continfo_pr2_so1_"+carnet),
		LoopInterval:         interval,
		MetricsAddr:          getenv("METRICS_ADDR", ":9105"),
		ValkeyAddr:           getenv("VALKEY_ADDR", "127.0.0.1:6379"),
		ValkeyPassword:       os.Getenv("VALKEY_PASSWORD"),
		ComposeFile:          absolute(absRoot, getenv("COMPOSE_FILE", "docker-compose.yml")),
		ProjectRoot:          absRoot,
		BPFObject:            absolute(absRoot, getenv("BPF_OBJECT", "daemon/internal/ebpf/kill_monitor.bpf.o")),
		MinLowContainers:     intEnv("MIN_LOW_CONTAINERS", 3),
		MinHighContainers:    intEnv("MIN_HIGH_CONTAINERS", 2),
		DeleteConfirmTimeout: timeout,
		EBPFRequired:         boolEnv("EBPF_REQUIRED", true),
	}
	if cfg.Carnet == "" || cfg.MinLowContainers < 0 || cfg.MinHighContainers < 0 {
		return Config{}, fmt.Errorf("invalid configuration")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func absolute(root, value string) string {
	if filepath.IsAbs(value) {
		return value
	}
	return filepath.Join(root, value)
}

func intEnv(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return value
}

func boolEnv(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}
