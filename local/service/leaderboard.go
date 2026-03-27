package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/logging"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type LeaderboardService struct {
	repo        *repository.LeaderboardRepository
	activityLog *ActivityLogService
}

func NewLeaderboardService(repo *repository.LeaderboardRepository, activityLog *ActivityLogService) *LeaderboardService {
	return &LeaderboardService{repo: repo, activityLog: activityLog}
}

// Get returns the leaderboard for a server.
func (s *LeaderboardService) Get(ctx context.Context, serverID uuid.UUID) (*model.Leaderboard, error) {
	return s.repo.GetOrCreateByServerID(ctx, serverID)
}

// Update replaces the leaderboard for a server.
func (s *LeaderboardService) Update(ctx context.Context, serverID uuid.UUID, lb *model.Leaderboard, actorID, actorUsername string) (*model.Leaderboard, error) {
	if lb == nil {
		return nil, fmt.Errorf("input is required")
	}
	for i := range lb.Drivers {
		lb.Drivers[i].Position = i
	}
	for i := range lb.Races {
		lb.Races[i].Position = i
	}
	result, err := s.repo.FullReplace(ctx, serverID, lb)
	if err != nil {
		return nil, err
	}

	driverCount := len(lb.Drivers)
	raceCount := len(lb.Races)
	details := fmt.Sprintf(`{"driver_count":%d,"race_count":%d}`, driverCount, raceCount)
	go func() {
		if logErr := s.activityLog.Log(context.Background(), &model.ActivityLog{
			ServerID: &serverID,
			UserID:   actorID,
			Username: actorUsername,
			Action:   model.ActionLeaderboardUpdate,
			Details:  details,
		}); logErr != nil {
			logging.Error("Failed to log leaderboard update: %v", logErr)
		}
	}()

	return result, nil
}
