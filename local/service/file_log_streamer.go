package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/utl/tracking"
	"bufio"
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
)

type FileLogStreamer struct {
	tailers sync.Map // uuid.UUID → *tracking.LogTailer
}

func NewFileLogStreamer() *FileLogStreamer {
	return &FileLogStreamer{}
}

func (s *FileLogStreamer) Start(_ context.Context, server *model.Server, handleLine func(string)) error {
	if _, exists := s.tailers.Load(server.ID); exists {
		return nil
	}

	logPath := filepath.Join(server.GetLogPath(), "server.log")
	tailer := tracking.NewLogTailer(logPath, handleLine)
	s.tailers.Store(server.ID, tailer)

	go tailer.Start()
	return nil
}

func (s *FileLogStreamer) Stop(serverID uuid.UUID) {
	if v, exists := s.tailers.LoadAndDelete(serverID); exists {
		v.(*tracking.LogTailer).Stop()
	}
}

func (s *FileLogStreamer) GetLastLines(_ context.Context, server *model.Server, n int) ([]string, error) {
	if n <= 0 {
		return []string{}, nil
	}
	logPath := filepath.Join(server.GetLogPath(), "server.log")
	f, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(lines) <= n {
		return lines, nil
	}
	return lines[len(lines)-n:], nil
}
