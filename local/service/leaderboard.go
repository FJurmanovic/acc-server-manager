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

// Get returns the leaderboard for a server in the wire format (LeaderboardInput).
func (s *LeaderboardService) Get(ctx context.Context, serverID uuid.UUID) (*model.LeaderboardInput, error) {
	lb, err := s.repo.GetOrCreateByServerID(ctx, serverID)
	if err != nil {
		return nil, err
	}
	return toInput(lb), nil
}

// Update replaces the leaderboard for a server from wire format input.
func (s *LeaderboardService) Update(ctx context.Context, serverID uuid.UUID, input *model.LeaderboardInput) (*model.LeaderboardInput, error) {
	lb, err := fromInput(serverID, input)
	if err != nil {
		return nil, err
	}
	result, err := s.repo.FullReplace(ctx, serverID, lb)
	if err != nil {
		return nil, err
	}
	return toInput(result), nil
}

// toInput converts the normalized DB model to the flat wire format.
func toInput(lb *model.Leaderboard) *model.LeaderboardInput {
	// Build driver map (id -> index in slice) for result ordering
	driverIndexByID := make(map[uuid.UUID]int, len(lb.Drivers))

	// Sort drivers by position
	drivers := make([]model.LeaderboardDriverInput, len(lb.Drivers))
	for i, d := range lb.Drivers {
		drivers[i] = model.LeaderboardDriverInput{
			Name:     d.Name,
			Color:    d.Color,
			Initials: d.Initials,
		}
		driverIndexByID[d.ID] = d.Position
	}
	// Re-index by position order
	driverByPos := make(map[int]model.LeaderboardDriver, len(lb.Drivers))
	for _, d := range lb.Drivers {
		driverByPos[d.Position] = d
	}
	driverCount := len(lb.Drivers)
	orderedDrivers := make([]model.LeaderboardDriver, driverCount)
	for i := 0; i < driverCount; i++ {
		if d, ok := driverByPos[i]; ok {
			orderedDrivers[i] = d
			drivers[i] = model.LeaderboardDriverInput{
				Name:     d.Name,
				Color:    d.Color,
				Initials: d.Initials,
			}
			driverIndexByID[d.ID] = i
		}
	}

	// Build tracks
	raceByPos := make(map[int]model.LeaderboardRace, len(lb.Races))
	for _, r := range lb.Races {
		raceByPos[r.Position] = r
	}
	tracks := make([]model.LeaderboardTrackInput, len(lb.Races))
	for i := 0; i < len(lb.Races); i++ {
		race, ok := raceByPos[i]
		if !ok {
			continue
		}
		results := make([]interface{}, driverCount)
		for j := range results {
			results[j] = "0"
		}
		for _, res := range race.Results {
			idx, found := driverIndexByID[res.DriverID]
			if found && idx < driverCount {
				results[idx] = res.Score
			}
		}
		tracks[i] = model.LeaderboardTrackInput{
			Name:               race.Name,
			Results:            results,
			FastestLapInitials: race.FastestLapInitials,
		}
	}

	// Build points table
	pointsTable := make([]model.LeaderboardPointRowInput, len(lb.PointRows))
	for i, p := range lb.PointRows {
		pointsTable[i] = model.LeaderboardPointRowInput{
			Points:    p.Points,
			Label:     p.Label,
			Color:     p.Color,
			TextColor: p.TextColor,
			Priority:  p.Priority,
		}
	}

	flColor := lb.FLColor
	if flColor == "" {
		flColor = "#8b5cf6"
	}
	flTextColor := lb.FLTextColor
	if flTextColor == "" {
		flTextColor = "#000000"
	}

	return &model.LeaderboardInput{
		Drivers:     drivers,
		PointsTable: pointsTable,
		FLPoints: model.LeaderboardFLInput{
			Points:    lb.FLPoints,
			Label:     "FL +1",
			Color:     flColor,
			TextColor: flTextColor,
			Priority:  len(pointsTable) + 1,
		},
		Tracks: tracks,
	}
}

// fromInput converts the wire format to the normalized DB model.
func fromInput(serverID uuid.UUID, input *model.LeaderboardInput) (*model.Leaderboard, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	lb := &model.Leaderboard{
		ServerID:    serverID,
		FLPoints:    input.FLPoints.Points,
		FLColor:     input.FLPoints.Color,
		FLTextColor: input.FLPoints.TextColor,
	}

	// Build drivers
	drivers := make([]model.LeaderboardDriver, len(input.Drivers))
	driverIDs := make([]uuid.UUID, len(input.Drivers))
	for i, d := range input.Drivers {
		id := uuid.New()
		driverIDs[i] = id
		drivers[i] = model.LeaderboardDriver{
			ID:       id,
			Name:     d.Name,
			Initials: d.Initials,
			Color:    d.Color,
			Position: i,
		}
	}
	lb.Drivers = drivers

	// Build point rows
	pointRows := make([]model.LeaderboardPointRow, len(input.PointsTable))
	for i, p := range input.PointsTable {
		pointRows[i] = model.LeaderboardPointRow{
			Label:     p.Label,
			Points:    p.Points,
			Color:     p.Color,
			TextColor: p.TextColor,
			Priority:  p.Priority,
		}
	}
	lb.PointRows = pointRows

	// Build races
	races := make([]model.LeaderboardRace, len(input.Tracks))
	for i, t := range input.Tracks {
		results := make([]model.LeaderboardResult, 0, len(t.Results))
		for j, score := range t.Results {
			if j >= len(driverIDs) {
				break
			}
			scoreStr := fmt.Sprintf("%v", score)
			results = append(results, model.LeaderboardResult{
				DriverID: driverIDs[j],
				Score:    scoreStr,
			})
		}
		races[i] = model.LeaderboardRace{
			Name:               t.Name,
			Position:           i,
			FastestLapInitials: t.FastestLapInitials,
			Results:            results,
		}
	}
	lb.Races = races

	return lb, nil
}
