package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/utl/logging"
	"bufio"
	"context"
	"io"
	"sync"

	"github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/google/uuid"
)

type DockerLogStreamer struct {
	client  *dockerclient.Client
	cancels sync.Map // uuid.UUID → context.CancelFunc
}

func NewDockerLogStreamer(client *dockerclient.Client) *DockerLogStreamer {
	return &DockerLogStreamer{client: client}
}

func (s *DockerLogStreamer) Start(ctx context.Context, server *model.Server, handleLine func(string)) error {
	if _, exists := s.cancels.Load(server.ID); exists {
		return nil
	}
	if server.ContainerID == "" {
		logging.Warn("DockerLogStreamer.Start: no container ID for server %s, skipping", server.ID)
		return nil
	}

	streamCtx, cancel := context.WithCancel(ctx)
	s.cancels.Store(server.ID, cancel)

	go func() {
		defer s.cancels.Delete(server.ID)

		rc, err := s.client.ContainerLogs(streamCtx, server.ContainerID, types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
			Timestamps: false,
		})
		if err != nil {
			logging.Error("DockerLogStreamer: failed to attach to container %s: %v", server.ContainerID[:12], err)
			return
		}
		defer rc.Close()

		pr, pw := io.Pipe()
		go func() {
			defer pw.Close()
			if _, err := stdcopy.StdCopy(pw, pw, rc); err != nil && err != io.EOF {
				logging.Warn("DockerLogStreamer: stdcopy error for %s: %v", server.ServiceName, err)
			}
		}()

		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			handleLine(scanner.Text())
		}
	}()

	return nil
}

func (s *DockerLogStreamer) Stop(serverID uuid.UUID) {
	if v, ok := s.cancels.LoadAndDelete(serverID); ok {
		v.(context.CancelFunc)()
	}
}
