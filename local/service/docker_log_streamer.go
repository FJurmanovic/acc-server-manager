package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/utl/logging"
	"bufio"
	"context"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
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
	s.Stop(server.ID)

	if server.ContainerID == "" {
		logging.Warn("DockerLogStreamer.Start: no container ID for server %s, skipping", server.ID)
		return nil
	}

	since := time.Now().UTC().Format(time.RFC3339)
	streamCtx, cancel := context.WithCancel(ctx)
	s.cancels.Store(server.ID, cancel)

	go func() {
		defer s.cancels.Delete(server.ID)

		rc, err := s.client.ContainerLogs(streamCtx, server.ContainerID, container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
			Timestamps: false,
			Since:      since,
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

func (s *DockerLogStreamer) GetLastLines(ctx context.Context, server *model.Server, n int) ([]string, error) {
	if n <= 0 || server.ContainerID == "" {
		return []string{}, nil
	}

	rc, err := s.client.ContainerLogs(ctx, server.ContainerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       strconv.Itoa(n),
	})
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		if _, err := stdcopy.StdCopy(pw, pw, rc); err != nil && err != io.EOF {
			logging.Warn("DockerLogStreamer.GetLastLines: stdcopy error for %s: %v", server.ServiceName, err)
		}
	}()

	var lines []string
	scanner := bufio.NewScanner(pr)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
