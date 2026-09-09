package decision

import (
	"testing"

	"proyecto2-so1-201801391/internal/model"
)

func metric(id, profile string, cpu float64) model.ContainerMetric {
	return model.ContainerMetric{
		Container:  model.Container{ID: id, Profile: profile, Running: true},
		CPUPercent: cpu, MemoryPercent: cpu, VSZKB: uint64(cpu * 10), RSSKB: uint64(cpu * 5),
	}
}

func TestSelectVictimsKeepsMinimums(t *testing.T) {
	metrics := []model.ContainerMetric{
		metric("l1", "low", 1), metric("l2", "low", 2), metric("l3", "low", 3), metric("l4", "low", 9),
		metric("h1", "high", 1), metric("h2", "high", 2), metric("h3", "high", 8),
		metric("i1", "intruder", 0),
	}
	victims := SelectVictims(metrics, 3, 2)
	if len(victims) != 3 {
		t.Fatalf("expected 3 victims, got %d", len(victims))
	}
	found := map[string]bool{}
	for _, victim := range victims {
		found[victim.Container.ID] = true
	}
	if !found["l4"] || !found["h3"] || !found["i1"] {
		t.Fatalf("unexpected victims: %#v", found)
	}
}
