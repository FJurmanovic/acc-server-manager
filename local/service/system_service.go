package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/platform"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/env"
	"acc-server-manager/local/utl/logging"
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
)

type SystemService struct {
	dockerClient  *dockerclient.Client
	repo          *repository.ServerRepository
	runtime       platform.ServerRuntime
	serverService *ServerService
}

func NewSystemService(
	dockerClient *dockerclient.Client,
	repo *repository.ServerRepository,
	runtime platform.ServerRuntime,
	serverService *ServerService,
) *SystemService {
	return &SystemService{
		dockerClient:  dockerClient,
		repo:          repo,
		runtime:       runtime,
		serverService: serverService,
	}
}

func (s *SystemService) MigrateImage(ctx context.Context, stopRunning bool) (*model.MigrationResult, error) {
	image := env.GetACCImage()

	servers, err := s.repo.GetAll(ctx, &model.ServerFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %v", err)
	}

	result := &model.MigrationResult{Image: image}

	for _, server := range *servers {
		statusStr, _ := s.runtime.Status(ctx, server.ID)
		isRunning := statusStr == model.StatusRunning.String()

		if isRunning && !stopRunning {
			result.Skipped = append(result.Skipped, model.ServerMigrationEntry{
				ID:         server.ID.String(),
				Name:       server.Name,
				WasRunning: true,
			})
			continue
		}

		entry := model.ServerMigrationEntry{
			ID:         server.ID.String(),
			Name:       server.Name,
			WasRunning: isRunning,
		}

		if isRunning {
			if _, err := s.runtime.Stop(ctx, server.ID); err != nil {
				result.Failed = append(result.Failed, model.ServerMigrationError{
					ID:    server.ID.String(),
					Name:  server.Name,
					Error: fmt.Sprintf("stop failed: %v", err),
				})
				continue
			}
		}

		if server.ContainerID != "" {
			if err := s.dockerClient.ContainerRemove(ctx, server.ContainerID, container.RemoveOptions{Force: true}); err != nil {
				logging.Warn("Failed to remove container %s for server %s: %v", server.ContainerID, server.ID, err)
			}
			server.ContainerID = ""
		}

		if err := s.runtime.Create(ctx, &server); err != nil {
			result.Failed = append(result.Failed, model.ServerMigrationError{
				ID:    server.ID.String(),
				Name:  server.Name,
				Error: fmt.Sprintf("recreate failed: %v", err),
			})
			continue
		}
		if err := s.repo.Update(ctx, &server); err != nil {
			logging.Warn("Failed to persist new ContainerID for server %s: %v", server.ID, err)
		}

		if isRunning {
			if _, err := s.runtime.Start(ctx, server.ID); err != nil {
				result.Failed = append(result.Failed, model.ServerMigrationError{
					ID:    server.ID.String(),
					Name:  server.Name,
					Error: fmt.Sprintf("restart failed: %v", err),
				})
				continue
			}
			s.serverService.StartAccServerRuntime(&server)
		}

		result.Migrated = append(result.Migrated, entry)
	}

	return result, nil
}
