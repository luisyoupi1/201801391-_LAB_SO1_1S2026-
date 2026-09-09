package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"proyecto2-so1-201801391/internal/model"
)

type peaks struct {
	ID, Name string
	Memory   float64
	CPU      float64
}

type Registry struct {
	mu          sync.RWMutex
	memory      model.Memory
	current     []model.ContainerMetric
	history     map[string]peaks
	deleted     uint64
	ebpfEvents  uint64
	lastSuccess time.Time
}

func New() *Registry { return &Registry{history: make(map[string]peaks)} }

func (r *Registry) Update(snapshot model.Snapshot, containers []model.ContainerMetric) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.memory = snapshot.Memory
	r.current = append([]model.ContainerMetric(nil), containers...)
	r.lastSuccess = time.Now()
	for _, item := range containers {
		peak := r.history[item.Container.ID]
		peak.ID, peak.Name = item.Container.ID, item.Container.Name
		if item.MemoryPercent > peak.Memory {
			peak.Memory = item.MemoryPercent
		}
		if item.CPUPercent > peak.CPU {
			peak.CPU = item.CPUPercent
		}
		r.history[item.Container.ID] = peak
	}
}

func (r *Registry) ObserveEBPF()     { r.mu.Lock(); r.ebpfEvents++; r.mu.Unlock() }
func (r *Registry) ObserveDeletion() { r.mu.Lock(); r.deleted++; r.mu.Unlock() }

func (r *Registry) Serve(ctx context.Context, addr string) error {
	server := &http.Server{Addr: addr, Handler: r, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()
	err := server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (r *Registry) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/metrics" {
		http.NotFound(writer, request)
		return
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	writer.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(writer, "so1_ram_total_bytes %d\n", r.memory.TotalKB*1024)
	fmt.Fprintf(writer, "so1_ram_free_bytes %d\n", r.memory.FreeKB*1024)
	fmt.Fprintf(writer, "so1_ram_used_bytes %d\n", r.memory.UsedKB*1024)
	fmt.Fprintf(writer, "so1_containers_deleted_total %d\n", r.deleted)
	fmt.Fprintf(writer, "so1_ebpf_kill_events_total %d\n", r.ebpfEvents)
	fmt.Fprintf(writer, "so1_last_success_timestamp_seconds %d\n", r.lastSuccess.Unix())
	for _, item := range r.current {
		labels := labels(item.Container.ID, item.Container.Name, item.Container.Profile)
		fmt.Fprintf(writer, "so1_container_memory_percent%s %s\n", labels, number(item.MemoryPercent))
		fmt.Fprintf(writer, "so1_container_cpu_percent%s %s\n", labels, number(item.CPUPercent))
		fmt.Fprintf(writer, "so1_container_rss_bytes%s %d\n", labels, item.RSSKB*1024)
		fmt.Fprintf(writer, "so1_container_vsz_bytes%s %d\n", labels, item.VSZKB*1024)
	}
	keys := make([]string, 0, len(r.history))
	for key := range r.history {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		peak := r.history[key]
		short := peak.ID
		if len(short) > 12 {
			short = short[:12]
		}
		label := fmt.Sprintf("{container_id=%q,name=%q}", escape(short), escape(peak.Name))
		fmt.Fprintf(writer, "so1_container_memory_peak_percent%s %s\n", label, number(peak.Memory))
		fmt.Fprintf(writer, "so1_container_cpu_peak_percent%s %s\n", label, number(peak.CPU))
	}
}

func labels(id, name, profile string) string {
	if len(id) > 12 {
		id = id[:12]
	}
	return fmt.Sprintf("{container_id=%q,name=%q,profile=%q}", escape(id), escape(name), escape(profile))
}

func escape(value string) string {
	return strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\"", "\\\"").Replace(value)
}

func number(value float64) string { return strconv.FormatFloat(value, 'f', 4, 64) }
