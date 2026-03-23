package repository

import (
	"acc-server-manager/local/model"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ActivityLogRepository struct {
	*BaseRepository[model.ActivityLog, model.ActivityLogFilter]
}

func NewActivityLogRepository(db *gorm.DB) *ActivityLogRepository {
	return &ActivityLogRepository{
		BaseRepository: NewBaseRepository[model.ActivityLog, model.ActivityLogFilter](db, model.ActivityLog{}),
	}
}

func (r *ActivityLogRepository) Insert(ctx context.Context, entry *model.ActivityLog) error {
	return r.BaseRepository.Insert(ctx, entry)
}

func (r *ActivityLogRepository) GetAll(ctx context.Context, filter *model.ActivityLogFilter) (*[]model.ActivityLog, error) {
	result := &[]model.ActivityLog{}
	query := r.BaseRepository.db.WithContext(ctx).
		Model(&model.ActivityLog{}).
		Preload("Server")

	query = filter.ApplyFilter(query)

	offset, limit := filter.Pagination()
	query = query.Offset(offset).Limit(limit)

	field, desc := filter.GetSorting()
	if desc {
		query = query.Order(field + " DESC")
	} else {
		query = query.Order(field)
	}

	if err := query.Find(result).Error; err != nil {
		return nil, fmt.Errorf("error getting activity logs: %w", err)
	}
	return result, nil
}

type lastActivityRow struct {
	ServerID  string           `gorm:"column:server_id"`
	CreatedAt time.Time        `gorm:"column:created_at"`
	Username  string           `gorm:"column:username"`
	Action    model.ActionType `gorm:"column:action"`
}

// GetLastActivityByServerIDs returns the most recent activity entry per server.
// Uses a single GROUP BY query with MAX(created_at); SQLite selects the other
// columns from the row that holds the max value — documented SQLite behaviour.
func (r *ActivityLogRepository) GetLastActivityByServerIDs(ctx context.Context, serverIDs []uuid.UUID) (map[uuid.UUID]*model.LastActivityInfo, error) {
	out := make(map[uuid.UUID]*model.LastActivityInfo)
	if len(serverIDs) == 0 {
		return out, nil
	}

	// Convert to string slice so the IN clause works with SQLite UUID strings.
	ids := make([]string, len(serverIDs))
	for i, id := range serverIDs {
		ids[i] = id.String()
	}

	var rows []lastActivityRow
	err := r.BaseRepository.db.WithContext(ctx).
		Model(&model.ActivityLog{}).
		Select("server_id, MAX(created_at) as created_at, username, action").
		Where("server_id IN ?", ids).
		Group("server_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		serverUUID, err := uuid.Parse(row.ServerID)
		if err != nil {
			continue
		}
		out[serverUUID] = &model.LastActivityInfo{
			UpdatedAt: row.CreatedAt,
			Username:  row.Username,
			Action:    row.Action,
		}
	}
	return out, nil
}
