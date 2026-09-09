package procfs

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	input := `{"carnet":"201801391","memory":{"total_kb":1000,"free_kb":400,"used_kb":600},"processes":[{"pid":7,"name":"demo","cmdline":"demo --run","vsz_kb":20,"rss_kb":10,"memory_percent":1.0,"cpu_percent":2.5}]}`
	snapshot, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Carnet != "201801391" || len(snapshot.Processes) != 1 {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}
}

func TestRejectZeroMemory(t *testing.T) {
	_, err := Parse(strings.NewReader(`{"memory":{"total_kb":0},"processes":[]}`))
	if err == nil {
		t.Fatal("expected validation error")
	}
}
