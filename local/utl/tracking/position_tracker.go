package tracking

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type LogPosition struct {
	LastPosition int64  `json:"last_position"`
	LastRead     string `json:"last_read"`
}

type PositionTracker struct {
	positionFile string
}

func NewPositionTracker(logPath string) *PositionTracker {
	dir := filepath.Dir(logPath)
	base := filepath.Base(logPath)
	positionFile := filepath.Join(dir, "."+base+".position")

	return &PositionTracker{
		positionFile: positionFile,
	}
}

func (t *PositionTracker) LoadPosition() (*LogPosition, error) {
	data, err := os.ReadFile(t.positionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &LogPosition{}, nil
		}
		return nil, err
	}

	var pos LogPosition
	if err := json.Unmarshal(data, &pos); err != nil {
		return nil, err
	}

	return &pos, nil
}

func (t *PositionTracker) SavePosition(pos *LogPosition) error {
	data, err := json.Marshal(pos)
	if err != nil {
		return err
	}

	return os.WriteFile(t.positionFile, data, 0644)
}
