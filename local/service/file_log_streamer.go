package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/utl/tracking"
	"context"
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
