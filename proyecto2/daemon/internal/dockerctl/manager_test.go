package dockerctl

import (
	"context"
	"os"
	"path/filepath"
	"proyecto2-so1-201801391/internal/model"
	"testing"
)

func TestRemoveAutoRemoval(t *testing.T) {
	for _, scenario := range []struct {
		name, script string
		wantErr      bool
	}{
		{"already removed", "if [ \"$1\" = rm ]; then echo 'No such container' >&2; exit 1; fi\nexit 0", false},
		{"removal in progress", "if [ \"$1\" = rm ]; then echo 'removal of container abc is already in progress' >&2; exit 1; fi\nif [ ! -f seen ]; then touch seen; echo abc; fi", false},
		{"daemon unavailable", "echo 'daemon unavailable' >&2; exit 1", true},
		{"permission error", "if [ \"$1\" = rm ]; then echo 'permission denied' >&2; exit 1; fi\necho abc", true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "docker"), []byte("#!/bin/sh\n"+scenario.script+"\n"), 0755); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
			m := New(dir, "201801391", "")
			err := m.Remove(context.Background(), model.Container{ID: "abc"})
			if (err != nil) != scenario.wantErr {
				t.Fatalf("error=%v wantErr=%v", err, scenario.wantErr)
			}
		})
	}
}
