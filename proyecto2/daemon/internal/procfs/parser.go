package procfs

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"proyecto2-so1-201801391/internal/model"
)

func Read(path string) (model.Snapshot, error) {
	file, err := os.Open(path)
	if err != nil {
		return model.Snapshot{}, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()
	return Parse(file)
}

func Parse(reader io.Reader) (model.Snapshot, error) {
	var snapshot model.Snapshot
	decoder := json.NewDecoder(io.LimitReader(reader, 64<<20))
	if err := decoder.Decode(&snapshot); err != nil {
		return model.Snapshot{}, fmt.Errorf("decode kernel snapshot: %w", err)
	}
	if snapshot.Memory.TotalKB == 0 {
		return model.Snapshot{}, fmt.Errorf("invalid snapshot: total memory is zero")
	}
	if snapshot.Processes == nil {
		snapshot.Processes = []model.Process{}
	}
	return snapshot, nil
}
