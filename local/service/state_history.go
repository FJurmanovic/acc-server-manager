package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/logging"
	"sync"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/sync/errgroup"
)

type StateHistoryService struct {
	repository *repository.StateHistoryRepository
}

func NewStateHistoryService(repository *repository.StateHistoryRepository) *StateHistoryService {
	return &StateHistoryService{repository: repository}
}

func (s *StateHistoryService) GetAll(ctx *fiber.Ctx, filter *model.StateHistoryFilter) (*[]model.StateHistory, error) {
	result, err := s.repository.GetAll(ctx.UserContext(), filter)
	if err != nil {
		logging.Error("Error getting state history: %v", err)
		return nil, err
	}
	return result, nil
}

func (s *StateHistoryService) Insert(ctx *fiber.Ctx, model *model.StateHistory) error {
	if err := s.repository.Insert(ctx.UserContext(), model); err != nil {
		logging.Error("Error inserting state history: %v", err)
		return err
	}
	return nil
}

func (s *StateHistoryService) GetLastSessionID(ctx *fiber.Ctx, serverID uint) (uint, error) {
	return s.repository.GetLastSessionID(ctx.UserContext(), serverID)
}

func (s *StateHistoryService) GetStatistics(ctx *fiber.Ctx, filter *model.StateHistoryFilter) (*model.StateHistoryStats, error) {
	stats := &model.StateHistoryStats{}
	var mu sync.Mutex

	eg, gCtx := errgroup.WithContext(ctx.UserContext())

	// Get Summary Stats (Peak/Avg Players, Total Sessions)
	eg.Go(func() error {
		summary, err := s.repository.GetSummaryStats(gCtx, filter)
		if err != nil {
			logging.Error("Error getting summary stats: %v", err)
			return err
		}
		mu.Lock()
		stats.PeakPlayers = summary.PeakPlayers
		stats.AveragePlayers = summary.AveragePlayers
		stats.TotalSessions = summary.TotalSessions
		mu.Unlock()
		return nil
	})

	// Get Total Playtime
	eg.Go(func() error {
		playtime, err := s.repository.GetTotalPlaytime(gCtx, filter)
		if err != nil {
			logging.Error("Error getting total playtime: %v", err)
			return err
		}
		mu.Lock()
		stats.TotalPlaytime = playtime
		mu.Unlock()
		return nil
	})

	// Get Player Count Over Time
	eg.Go(func() error {
		playerCount, err := s.repository.GetPlayerCountOverTime(gCtx, filter)
		if err != nil {
			logging.Error("Error getting player count over time: %v", err)
			return err
		}
		mu.Lock()
		stats.PlayerCountOverTime = playerCount
		mu.Unlock()
		return nil
	})

	// Get Session Types
	eg.Go(func() error {
		sessionTypes, err := s.repository.GetSessionTypes(gCtx, filter)
		if err != nil {
			logging.Error("Error getting session types: %v", err)
			return err
		}
		mu.Lock()
		stats.SessionTypes = sessionTypes
		mu.Unlock()
		return nil
	})

	// Get Daily Activity
	eg.Go(func() error {
		dailyActivity, err := s.repository.GetDailyActivity(gCtx, filter)
		if err != nil {
			logging.Error("Error getting daily activity: %v", err)
			return err
		}
		mu.Lock()
		stats.DailyActivity = dailyActivity
		mu.Unlock()
		return nil
	})

	// Get Recent Sessions
	eg.Go(func() error {
		recentSessions, err := s.repository.GetRecentSessions(gCtx, filter)
		if err != nil {
			logging.Error("Error getting recent sessions: %v", err)
			return err
		}
		mu.Lock()
		stats.RecentSessions = recentSessions
		mu.Unlock()
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return stats, nil
}