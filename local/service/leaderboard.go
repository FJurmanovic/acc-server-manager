package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type LeaderboardService struct {
	repo *repository.LeaderboardRepository
}

func NewLeaderboardService(repo *repository.LeaderboardRepository) *LeaderboardService {
	return &LeaderboardService{repo: repo}
}

// Get returns the leaderboard for a server.
func (s *LeaderboardService) Get(ctx context.Context, serverID uuid.UUID) (*model.Leaderboard, error) {
	return s.repo.GetOrCreateByServerID(ctx, serverID)
}

// Update replaces the leaderboard for a server.
func (s *LeaderboardService) Update(ctx context.Context, serverID uuid.UUID, lb *model.Leaderboard) (*model.Leaderboard, error) {
	if lb == nil {
		return nil, fmt.Errorf("input is required")
	}
	for i := range lb.Drivers {
		lb.Drivers[i].Position = i
	}
	for i := range lb.Races {
		lb.Races[i].Position = i
	}
	return s.repo.FullReplace(ctx, serverID, lb)
}
