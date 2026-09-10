package metrics

import (
	"net/http/httptest"
	"proyecto2-so1-201801391/internal/model"
	"strings"
	"testing"
)

func TestLabelsEscapeOnce(t *testing.T) {
	got := labels("123", "a\"b\\c\nd", "low")
	want := `{container_id="123",name="a\"b\\c\nd",profile="low"}`
	if got != want {
		t.Fatalf("got %s; want %s", got, want)
	}
}

func TestMetricsLifecycle(t *testing.T) {
	r := New()
	request := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, request)
	if !strings.Contains(rec.Body.String(), "so1_last_success_timestamp_seconds 0\n") {
		t.Fatal(rec.Body.String())
	}
	r.Update(model.Snapshot{}, []model.ContainerMetric{
		{Container: model.Container{ID: "live", Name: "live", Running: true}, MemoryPercent: 12},
		{Container: model.Container{ID: "stopped", Name: "stopped", Running: false}, MemoryPercent: 5},
	})
	r.ObserveDeletion()
	r.ObserveEBPF()
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, request)
	body := rec.Body.String()
	for _, want := range []string{"so1_containers_deleted_total 1\n", "so1_ebpf_kill_events_total 1\n", `so1_container_memory_percent{container_id="live"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %s", want)
		}
	}
	if strings.Contains(body, `so1_container_memory_percent{container_id="stopped"`) {
		t.Fatal("stopped container exported as current")
	}
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/missing", nil))
	if rec.Code != 404 {
		t.Fatal(rec.Code)
	}
}
