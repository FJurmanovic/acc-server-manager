package repository

import (
	"acc-server-manager/local/model"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StateHistoryRepository struct {
	*BaseRepository[model.StateHistory, model.StateHistoryFilter]
}

func NewStateHistoryRepository(db *gorm.DB) *StateHistoryRepository {
	return &StateHistoryRepository{
		BaseRepository: NewBaseRepository[model.StateHistory, model.StateHistoryFilter](db, model.StateHistory{}),
	}
}

func (r *StateHistoryRepository) GetAll(ctx context.Context, filter *model.StateHistoryFilter) (*[]model.StateHistory, error) {
	return r.BaseRepository.GetAll(ctx, filter)
}

func (r *StateHistoryRepository) Insert(ctx context.Context, model *model.StateHistory) error {
	return r.BaseRepository.Insert(ctx, model)
}

func (r *StateHistoryRepository) GetLastSessionID(ctx context.Context, serverID uuid.UUID) (uuid.UUID, error) {
	var lastSession model.StateHistory
	result := r.BaseRepository.db.WithContext(ctx).
		Where("server_id = ?", serverID).
		Order("session_id DESC").
		First(&lastSession)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return uuid.Nil, nil
		}
		return uuid.Nil, result.Error
	}

	return lastSession.SessionID, nil
}

func (r *StateHistoryRepository) GetSummaryStats(ctx context.Context, filter *model.StateHistoryFilter) (model.StateHistoryStats, error) {
	var stats model.StateHistoryStats
	serverUUID, err := uuid.Parse(filter.ServerID)
	if err != nil {
		return model.StateHistoryStats{}, err
	}

	query := r.db.WithContext(ctx).Model(&model.StateHistory{}).
		Select(`
			COALESCE(MAX(player_count), 0) as peak_players,
			COUNT(DISTINCT session_id) as total_sessions,
			COALESCE(AVG(player_count), 0) as average_players
		`).
		Where("server_id = ?", serverUUID)

	if !filter.StartDate.IsZero() && !filter.EndDate.IsZero() {
		query = query.Where("date_created BETWEEN ? AND ?", filter.StartDate, filter.EndDate)
	}

	if err := query.Scan(&stats).Error; err != nil {
		return model.StateHistoryStats{}, err
	}
	return stats, nil
}

func (r *StateHistoryRepository) GetTotalPlaytime(ctx context.Context, filter *model.StateHistoryFilter) (int, error) {
	var totalPlaytime struct {
		TotalMinutes float64
	}
	serverUUID, err := uuid.Parse(filter.ServerID)
	if err != nil {
		return 0, err
	}

	rawQuery := `
		SELECT SUM(duration_minutes) as total_minutes FROM (
			SELECT (strftime('%s', MAX(date_created)) - strftime('%s', MIN(date_created))) / 60.0 as duration_minutes
			FROM state_histories
			WHERE server_id = ? AND date_created BETWEEN ? AND ?
			GROUP BY session_id
			HAVING COUNT(*) > 1 AND MAX(player_count) > 0
		)
	`
	err = r.db.WithContext(ctx).Raw(rawQuery, serverUUID, filter.StartDate, filter.EndDate).Scan(&totalPlaytime).Error
	if err != nil {
		return 0, err
	}
	return int(totalPlaytime.TotalMinutes), nil
}

func (r *StateHistoryRepository) GetPlayerCountOverTime(ctx context.Context, filter *model.StateHistoryFilter) ([]model.PlayerCountPoint, error) {
	var points []model.PlayerCountPoint
	serverUUID, err := uuid.Parse(filter.ServerID)
	if err != nil {
		return points, err
	}

	rawQuery := `
		SELECT
			DATETIME(MIN(date_created)) as timestamp,
			ROUND(AVG(player_count)) as count
		FROM state_histories
		WHERE server_id = ? AND date_created BETWEEN ? AND ?
		GROUP BY strftime('%Y-%m-%d %H', date_created)
		ORDER BY timestamp
	`
	err = r.db.WithContext(ctx).Raw(rawQuery, serverUUID, filter.StartDate, filter.EndDate).Scan(&points).Error
	return points, err
}

func (r *StateHistoryRepository) GetSessionTypes(ctx context.Context, filter *model.StateHistoryFilter) ([]model.SessionCount, error) {
	var sessionTypes []model.SessionCount
	serverUUID, err := uuid.Parse(filter.ServerID)
	if err != nil {
		return sessionTypes, err
	}

	rawQuery := `
		SELECT session as name, COUNT(*) as count FROM (
			SELECT session
			FROM state_histories
			WHERE server_id = ? AND date_created BETWEEN ? AND ?
			GROUP BY session_id
		)
		GROUP BY session
		ORDER BY count DESC
	`
	err = r.db.WithContext(ctx).Raw(rawQuery, serverUUID, filter.StartDate, filter.EndDate).Scan(&sessionTypes).Error
	return sessionTypes, err
}

func (r *StateHistoryRepository) GetDailyActivity(ctx context.Context, filter *model.StateHistoryFilter) ([]model.DailyActivity, error) {
	var dailyActivity []model.DailyActivity
	serverUUID, err := uuid.Parse(filter.ServerID)
	if err != nil {
		return dailyActivity, err
	}

	rawQuery := `
		SELECT
			strftime('%Y-%m-%d', date_created) as date,
			COUNT(DISTINCT session_id) as sessions_count
		FROM state_histories
		WHERE server_id = ? AND date_created BETWEEN ? AND ?
		GROUP BY 1
		ORDER BY 1
	`
	err = r.db.WithContext(ctx).Raw(rawQuery, serverUUID, filter.StartDate, filter.EndDate).Scan(&dailyActivity).Error
	return dailyActivity, err
}

func (r *StateHistoryRepository) GetRecentSessions(ctx context.Context, filter *model.StateHistoryFilter) ([]model.RecentSession, error) {
	var recentSessions []model.RecentSession
	serverUUID, err := uuid.Parse(filter.ServerID)
	if err != nil {
		return recentSessions, err
	}

	rawQuery := `
		SELECT
			session_id as id,
			DATETIME(MIN(date_created)) as date,
			session as type,
			track,
			MAX(player_count) as players,
			CAST((strftime('%s', MAX(date_created)) - strftime('%s', MIN(date_created))) / 60 AS INTEGER) as duration
		FROM state_histories
		WHERE server_id = ? AND date_created BETWEEN ? AND ?
		GROUP BY session_id
		HAVING MAX(player_count) > 0
		ORDER BY date DESC
		LIMIT 10
	`
	err = r.db.WithContext(ctx).Raw(rawQuery, serverUUID, filter.StartDate, filter.EndDate).Scan(&recentSessions).Error
	return recentSessions, err
}
