package config

import "testing"

func TestRejectNonpositiveDurations(t *testing.T) {
	for _, key := range []string{"LOOP_INTERVAL", "DELETE_CONFIRM_TIMEOUT"} {
		for _, value := range []string{"0s", "-1s", "invalid"} {
			t.Run(key+value, func(t *testing.T) {
				t.Setenv(key, value)
				if _, err := Load(); err == nil {
					t.Fatal("expected invalid duration to be rejected")
				}
			})
		}
	}
}

func TestValidDuration(t *testing.T) {
	t.Setenv("LOOP_INTERVAL", "2s")
	t.Setenv("DELETE_CONFIRM_TIMEOUT", "1s")
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
}
