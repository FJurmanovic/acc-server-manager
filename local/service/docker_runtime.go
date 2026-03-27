package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/env"
	"acc-server-manager/local/utl/logging"
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
)

type DockerRuntime struct {
	client *dockerclient.Client
	repo   *repository.ServerRepository
}

func NewDockerRuntime(client *dockerclient.Client, repo *repository.ServerRepository) *DockerRuntime {
	return &DockerRuntime{
		client: client,
		repo:   repo,
	}
}

func (r *DockerRuntime) Create(ctx context.Context, server *model.Server) error {
	image := env.GetACCImage()
	// ACC uses the same port number for both TCP (client connections) and UDP (car positions/ping).
	portStr := strconv.Itoa(server.Port)

	natTCP, err := nat.NewPort("tcp", portStr)
	if err != nil {
		return fmt.Errorf("invalid port %d: %v", server.Port, err)
	}
	natUDP, err := nat.NewPort("udp", portStr)
	if err != nil {
		return fmt.Errorf("invalid port %d: %v", server.Port, err)
	}

	portBindings := nat.PortMap{
		natTCP: []nat.PortBinding{{HostPort: portStr}},
		natUDP: []nat.PortBinding{{HostPort: portStr}},
	}
	exposedPorts := nat.PortSet{
		natTCP: struct{}{},
		natUDP: struct{}{},
	}

	containerBase := env.GetACCServersPath()
	hostBase := env.GetACCServersHostPath()
	serverSubPath := strings.TrimPrefix(server.GetServerPath(), containerBase)
	hostServerPath := filepath.Join(hostBase, serverSubPath)

	binds := []string{
		fmt.Sprintf("%s:/acc/game:rw", hostServerPath),
	}

	resp, err := r.client.ContainerCreate(ctx, &container.Config{
		Image:        image,
		ExposedPorts: exposedPorts,
		Env: []string{
			fmt.Sprintf("SERVER_PORT=%d", server.Port),
		},
	}, &container.HostConfig{
		Binds:         binds,
		PortBindings:  portBindings,
		RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
		// NET_RAW is required for tcpdump packet capture in debug mode (DEBUG_MODE=1).
		CapAdd: []string{"NET_RAW"},
	}, nil, nil, server.ServiceName)
	if err != nil {
		return fmt.Errorf("failed to create container %s: %v", server.ServiceName, err)
	}

	server.ContainerID = resp.ID
	logging.Info("Created Docker container %s (ID: %s) for server %s", server.ServiceName, resp.ID[:12], server.ID)
	return nil
}

func (r *DockerRuntime) Delete(ctx context.Context, serverID uuid.UUID) error {
	containerID, err := r.resolveContainerID(ctx, serverID)
	if err != nil {
		return err
	}
	if containerID == "" {
		return nil
	}

	err = r.client.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force:         true,
		RemoveVolumes: false,
	})
	if err != nil {
		return fmt.Errorf("failed to remove container %s: %v", containerID[:12], err)
	}
	logging.Info("Removed Docker container %s for server %s", containerID[:12], serverID)
	return nil
}

func (r *DockerRuntime) Start(ctx context.Context, serverID uuid.UUID) (string, error) {
	server, err := r.repo.GetByID(ctx, serverID)
	if err != nil {
		return "", fmt.Errorf("failed to look up server %s: %v", serverID, err)
	}
	if server == nil {
		return "", fmt.Errorf("server %s not found", serverID)
	}

	containerID, err := r.ensureContainer(ctx, server)
	if err != nil {
		return "", err
	}

	if err := r.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %v", err)
	}
	return model.StatusRunning.String(), nil
}

func (r *DockerRuntime) Stop(ctx context.Context, serverID uuid.UUID) (string, error) {
	containerID, err := r.resolveContainerID(ctx, serverID)
	if err != nil {
		return "", err
	}
	if err := r.client.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
		return "", fmt.Errorf("failed to stop container: %v", err)
	}
	return model.StatusStopped.String(), nil
}

func (r *DockerRuntime) Restart(ctx context.Context, serverID uuid.UUID) (string, error) {
	server, err := r.repo.GetByID(ctx, serverID)
	if err != nil {
		return "", fmt.Errorf("failed to look up server %s: %v", serverID, err)
	}
	if server == nil {
		return "", fmt.Errorf("server %s not found", serverID)
	}

	containerID, err := r.ensureContainer(ctx, server)
	if err != nil {
		return "", err
	}

	if err := r.client.ContainerRestart(ctx, containerID, container.StopOptions{}); err != nil {
		return "", fmt.Errorf("failed to restart container: %v", err)
	}
	return model.StatusRunning.String(), nil
}

// ensureContainer returns the container ID for a server, recreating the container
// from the server's current DB record if it no longer exists in Docker.
func (r *DockerRuntime) ensureContainer(ctx context.Context, server *model.Server) (string, error) {
	if server.ContainerID != "" {
		if _, err := r.client.ContainerInspect(ctx, server.ContainerID); err == nil {
			return server.ContainerID, nil
		}
		logging.Info("Container %s for server %s no longer exists, recreating", server.ContainerID[:min(12, len(server.ContainerID))], server.ID)
	}

	if err := r.Create(ctx, server); err != nil {
		return "", fmt.Errorf("failed to recreate container: %v", err)
	}
	if err := r.repo.Update(ctx, server); err != nil {
		return "", fmt.Errorf("container recreated but failed to persist container ID: %v", err)
	}
	return server.ContainerID, nil
}

func (r *DockerRuntime) Status(ctx context.Context, serverID uuid.UUID) (string, error) {
	server, err := r.repo.GetByID(ctx, serverID)
	if err != nil || server == nil {
		return model.StatusUnknown.String(), err
	}

	containerID := server.ContainerID
	if containerID == "" {
		containerID, err = r.findContainerByName(ctx, server.ServiceName)
		if err != nil || containerID == "" {
			return model.StatusStopped.String(), nil
		}
	}

	info, err := r.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return model.StatusUnknown.String(), nil
	}

	return dockerStateToStatus(info.State.Status), nil
}

func (r *DockerRuntime) resolveContainerID(ctx context.Context, serverID uuid.UUID) (string, error) {
	server, err := r.repo.GetByID(ctx, serverID)
	if err != nil {
		return "", fmt.Errorf("failed to look up server %s: %v", serverID, err)
	}
	if server == nil {
		return "", fmt.Errorf("server %s not found", serverID)
	}
	return server.ContainerID, nil
}

func (r *DockerRuntime) findContainerByName(ctx context.Context, name string) (string, error) {
	containers, err := r.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("name", "/"+name)),
	})
	if err != nil {
		return "", err
	}
	if len(containers) == 0 {
		return "", nil
	}
	return containers[0].ID, nil
}

func dockerStateToStatus(state string) string {
	switch state {
	case "running":
		return model.StatusRunning.String()
	case "exited", "dead":
		return model.StatusStopped.String()
	case "restarting":
		return model.StatusRestarting.String()
	case "paused":
		return model.StatusStopped.String()
	default:
		return model.StatusUnknown.String()
	}
}
