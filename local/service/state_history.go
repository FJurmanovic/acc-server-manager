package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/logging"
	"sort"
	"time"

	"github.com/gofiber/fiber/v2"
)

type StateHistoryService struct {
	repository *repository.StateHistoryRepository
}

func NewStateHistoryService(repository *repository.StateHistoryRepository) *StateHistoryService {
	return &StateHistoryService{
		repository: repository,
	}
}

// GetAll
// Gets All rows from StateHistory table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			string: Application version
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

func (s *StateHistoryService) GetStatistics(ctx *fiber.Ctx, filter *model.StateHistoryFilter) (*model.StateHistoryStats, error) {
	// Get all state history entries based on filter
	entries, err := s.repository.GetAll(ctx.UserContext(), filter)
	if err != nil {
		logging.Error("Error getting state history for statistics: %v", err)
		return nil, err
	}

	stats := &model.StateHistoryStats{
		PlayerCountOverTime: make([]model.PlayerCountPoint, 0),
		SessionTypes:        make([]model.SessionCount, 0),
		DailyActivity:      make([]model.DailyActivity, 0),
		RecentSessions:     make([]model.RecentSession, 0),
	}

	if len(*entries) == 0 {
		return stats, nil
	}

	// Maps to track unique sessions and their details
	sessionMap := make(map[uint]*struct {
		StartTime  time.Time
		EndTime    time.Time
		Session    string
		Track      string
		MaxPlayers int
		SessionConcluded bool
	})

	// Maps for aggregating statistics
	dailySessionCount := make(map[string]int)
	sessionTypeCount := make(map[string]int)
	totalPlayers := 0
	peakPlayers := 0

	// Process each state history entry
	for _, entry := range *entries {
		// Track player count over time
		stats.PlayerCountOverTime = append(stats.PlayerCountOverTime, model.PlayerCountPoint{
			Timestamp: entry.DateCreated,
			Count:     entry.PlayerCount,
		})

		// Update peak players
		if entry.PlayerCount > peakPlayers {
			peakPlayers = entry.PlayerCount
		}

		totalPlayers += entry.PlayerCount

		// Process session information using SessionID
		if _, exists := sessionMap[entry.SessionID]; !exists {
			sessionMap[entry.SessionID] = &struct {
				StartTime  time.Time
				EndTime    time.Time
				Session    string
				Track      string
				MaxPlayers int
				SessionConcluded bool
			}{
				StartTime:  entry.DateCreated,
				Session:    entry.Session,
				Track:      entry.Track,
				MaxPlayers: entry.PlayerCount,
				SessionConcluded: false,
			}

			// Count session types
			sessionTypeCount[entry.Session]++

			// Count daily sessions
			dateStr := entry.DateCreated.Format("2006-01-02")
			dailySessionCount[dateStr]++
		} else {
			session := sessionMap[entry.SessionID]
			if session.SessionConcluded {
				continue
			}
			if (entry.PlayerCount == 0) {
				session.SessionConcluded = true
			}
			session.EndTime = entry.DateCreated
			if entry.PlayerCount > session.MaxPlayers {
				session.MaxPlayers = entry.PlayerCount
			}
		}
	}

	for key, session := range sessionMap {
		if !session.SessionConcluded {
			session.SessionConcluded = true
		}
		if (session.MaxPlayers == 0) {
			delete(sessionMap, key)
		}
	}

	// Calculate statistics
	stats.PeakPlayers = peakPlayers
	stats.TotalSessions = len(sessionMap)
	if len(*entries) > 0 {
		stats.AveragePlayers = float64(totalPlayers) / float64(len(*entries))
	}

	// Process session types
	for sessionType, count := range sessionTypeCount {
		stats.SessionTypes = append(stats.SessionTypes, model.SessionCount{
			Name:  sessionType,
			Count: count,
		})
	}

	// Process daily activity
	for dateStr, count := range dailySessionCount {
		date, _ := time.Parse("2006-01-02", dateStr)
		stats.DailyActivity = append(stats.DailyActivity, model.DailyActivity{
			Date:          date,
			SessionsCount: count,
		})
	}

	// Calculate total playtime and prepare recent sessions
	var recentSessions []model.RecentSession
	totalPlaytime := 0

	for sessionID, session := range sessionMap {
		if !session.EndTime.IsZero() {
			duration := int(session.EndTime.Sub(session.StartTime).Minutes())
			totalPlaytime += duration

			recentSessions = append(recentSessions, model.RecentSession{
				ID:       sessionID,
				Date:     session.StartTime,
				Type:     session.Session,
				Track:    session.Track,
				Duration: duration,
				Players:  session.MaxPlayers,
			})
		}
	}

	stats.TotalPlaytime = totalPlaytime

	// Sort recent sessions by date (newest first) and limit to last 10
	sort.Slice(recentSessions, func(i, j int) bool {
		return recentSessions[i].Date.After(recentSessions[j].Date)
	})

	if len(recentSessions) > 10 {
		recentSessions = recentSessions[:10]
	}
	stats.RecentSessions = recentSessions

	// Sort daily activity by date
	sort.Slice(stats.DailyActivity, func(i, j int) bool {
		return stats.DailyActivity[i].Date.Before(stats.DailyActivity[j].Date)
	})

	// Sort player count over time by timestamp
	sort.Slice(stats.PlayerCountOverTime, func(i, j int) bool {
		return stats.PlayerCountOverTime[i].Timestamp.Before(stats.PlayerCountOverTime[j].Timestamp)
	})

	return stats, nil
}