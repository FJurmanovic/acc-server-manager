package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"context"
	"path/filepath"

	"github.com/google/uuid"
)

type WindowsRuntime struct {
	windowsService *WindowsService
	repo           *repository.ServerRepository
}

func NewWindowsRuntime(windowsService *WindowsService, repo *repository.ServerRepository) *WindowsRuntime {
	return &WindowsRuntime{
		windowsService: windowsService,
		repo:           repo,
	}
}

func (r *WindowsRuntime) Create(ctx context.Context, server *model.Server) error {
	execPath := filepath.Join(server.GetServerPath(), "accServer.exe")
	serverWorkingDir := server.GetServerPath()
	return r.windowsService.CreateService(ctx, server.ServiceName, execPath, serverWorkingDir, nil)
}

func (r *WindowsRuntime) Delete(ctx context.Context, serverID uuid.UUID) error {
	server, err := r.repo.GetByID(ctx, serverID)
	if err != nil {
		return err
	}
	return r.windowsService.DeleteService(ctx, server.ServiceName)
}

func (r *WindowsRuntime) Start(ctx context.Context, serverID uuid.UUID) (string, error) {
	server, err := r.repo.GetByID(ctx, serverID)
	if err != nil {
		return "", err
	}
	return r.windowsService.Start(ctx, server.ServiceName)
}

func (r *WindowsRuntime) Stop(ctx context.Context, serverID uuid.UUID) (string, error) {
	server, err := r.repo.GetByID(ctx, serverID)
	if err != nil {
		return "", err
	}
	return r.windowsService.Stop(ctx, server.ServiceName)
}

func (r *WindowsRuntime) Restart(ctx context.Context, serverID uuid.UUID) (string, error) {
	server, err := r.repo.GetByID(ctx, serverID)
	if err != nil {
		return "", err
	}
	return r.windowsService.Restart(ctx, server.ServiceName)
}

func (r *WindowsRuntime) Status(ctx context.Context, serverID uuid.UUID) (string, error) {
	server, err := r.repo.GetByID(ctx, serverID)
	if err != nil {
		return "", err
	}
	return r.windowsService.Status(ctx, server.ServiceName)
}
