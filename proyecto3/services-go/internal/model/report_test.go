package model

import (
	"testing"
	"time"
)

func TestValidate(t *testing.T) {
	valid := WarReport{ID: "1", Country: "USA", WarplanesInAir: 50, WarshipsInWater: 30, Timestamp: time.Now()}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid report rejected: %v", err)
	}
	invalid := valid
	invalid.Country = "XXX"
	if invalid.Validate() == nil {
		t.Fatal("invalid country accepted")
	}
}
