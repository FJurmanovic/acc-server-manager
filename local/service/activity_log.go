package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/logging"
	"context"
	"encoding/json"
	"fmt"
)

type ActivityLogService struct {
	repo *repository.ActivityLogRepository
}

func NewActivityLogService(repo *repository.ActivityLogRepository) *ActivityLogService {
	return &ActivityLogService{repo: repo}
}

// Log persists an activity log entry. Errors are logged but not returned so
// callers can treat this as fire-and-forget.
func (s *ActivityLogService) Log(ctx context.Context, entry *model.ActivityLog) error {
	if err := s.repo.Insert(ctx, entry); err != nil {
		logging.Error("Failed to insert activity log: %v", err)
		return err
	}
	return nil
}

func (s *ActivityLogService) GetAll(ctx context.Context, filter *model.ActivityLogFilter) (*[]model.ActivityLog, error) {
	return s.repo.GetAll(ctx, filter)
}

// BuildConfigDetails produces a JSON string summarising which top-level keys
// changed between oldJSON and newJSON.
func BuildConfigDetails(file string, oldJSON, newJSON []byte) string {
	var oldMap, newMap map[string]interface{}

	if err := json.Unmarshal(oldJSON, &oldMap); err != nil {
		oldMap = map[string]interface{}{}
	}
	if err := json.Unmarshal(newJSON, &newMap); err != nil {
		newMap = map[string]interface{}{}
	}

	var changedKeys []string
	for key, newVal := range newMap {
		oldVal, exists := oldMap[key]
		if !exists {
			changedKeys = append(changedKeys, key)
			continue
		}
		oldBytes, _ := json.Marshal(oldVal)
		newBytes, _ := json.Marshal(newVal)
		if string(oldBytes) != string(newBytes) {
			changedKeys = append(changedKeys, key)
		}
	}

	if len(changedKeys) == 0 {
		return ""
	}
	keysJSON, _ := json.Marshal(changedKeys)
	return fmt.Sprintf(`{"file":%q,"changed_keys":%s}`, file, string(keysJSON))
}
