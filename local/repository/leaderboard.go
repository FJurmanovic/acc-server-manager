package repository

import (
	"acc-server-manager/local/model"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LeaderboardRepository struct {
	db *gorm.DB
}

func NewLeaderboardRepository(db *gorm.DB) *LeaderboardRepository {
	return &LeaderboardRepository{db: db}
}

// GetOrCreateByServerID fetches the leaderboard for a server, creating an empty one if it doesn't exist.
func (r *LeaderboardRepository) GetOrCreateByServerID(ctx context.Context, serverID uuid.UUID) (*model.Leaderboard, error) {
	lb := new(model.Leaderboard)
	err := r.db.WithContext(ctx).
		Preload("Drivers").
		Preload("Races.Results").
		Preload("PointRows").
		Where("server_id = ?", serverID).
		First(lb).Error

	if err == nil {
		return lb, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("error fetching leaderboard: %w", err)
	}

	lb = &model.Leaderboard{
		ServerID:    serverID,
		FLPoints:    1,
		FLColor:     "#8b5cf6",
		FLTextColor: "#000000",
	}
	if err := r.db.WithContext(ctx).Create(lb).Error; err != nil {
		return nil, fmt.Errorf("error creating leaderboard: %w", err)
	}
	return lb, nil
}

// FullReplace replaces all leaderboard data for a server in a single transaction.
func (r *LeaderboardRepository) FullReplace(ctx context.Context, serverID uuid.UUID, lb *model.Leaderboard) (*model.Leaderboard, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		existing := new(model.Leaderboard)
		err := tx.Where("server_id = ?", serverID).First(existing).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("error fetching existing leaderboard: %w", err)
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			lb.ServerID = serverID
			return tx.Create(lb).Error
		}

		// Delete children
		if err := tx.Where("leaderboard_id = ?", existing.ID).Delete(&model.LeaderboardPointRow{}).Error; err != nil {
			return err
		}
		if err := tx.Where("leaderboard_id = ?", existing.ID).Delete(&model.LeaderboardDriver{}).Error; err != nil {
			return err
		}

		// Get race IDs to delete results
		var raceIDs []uuid.UUID
		tx.Model(&model.LeaderboardRace{}).Where("leaderboard_id = ?", existing.ID).Pluck("id", &raceIDs)
		if len(raceIDs) > 0 {
			if err := tx.Where("race_id IN ?", raceIDs).Delete(&model.LeaderboardResult{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("leaderboard_id = ?", existing.ID).Delete(&model.LeaderboardRace{}).Error; err != nil {
			return err
		}

		// Update parent fields
		existing.FLPoints = lb.FLPoints
		existing.FLColor = lb.FLColor
		existing.FLTextColor = lb.FLTextColor
		if err := tx.Save(existing).Error; err != nil {
			return err
		}

		// Re-insert children with correct leaderboard ID
		for i := range lb.Drivers {
			lb.Drivers[i].LeaderboardID = existing.ID
			lb.Drivers[i].ID = uuid.Nil // force new ID
		}
		if len(lb.Drivers) > 0 {
			if err := tx.Create(&lb.Drivers).Error; err != nil {
				return err
			}
		}

		for i := range lb.PointRows {
			lb.PointRows[i].LeaderboardID = existing.ID
			lb.PointRows[i].ID = uuid.Nil
		}
		if len(lb.PointRows) > 0 {
			if err := tx.Create(&lb.PointRows).Error; err != nil {
				return err
			}
		}

		for i := range lb.Races {
			lb.Races[i].LeaderboardID = existing.ID
			lb.Races[i].ID = uuid.Nil
		}
		if len(lb.Races) > 0 {
			// Create races first (without results) to get IDs
			for i := range lb.Races {
				results := lb.Races[i].Results
				lb.Races[i].Results = nil
				if err := tx.Create(&lb.Races[i]).Error; err != nil {
					return err
				}
				// Re-attach results with correct race ID
				for j := range results {
					results[j].RaceID = lb.Races[i].ID
					results[j].ID = uuid.Nil
				}
				if len(results) > 0 {
					if err := tx.Create(&results).Error; err != nil {
						return err
					}
				}
			}
		}

		lb.ID = existing.ID
		return nil
	})
	if err != nil {
		return nil, err
	}

	return r.GetOrCreateByServerID(ctx, serverID)
}
