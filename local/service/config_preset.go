package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

const (
	presetMaxNameLen        = 100
	presetMaxDescLen        = 500
	presetMaxSectionBytes   = 64 * 1024 // 64 KB per config section
)

type ConfigPresetService struct {
	repository    *repository.ConfigPresetRepository
	configService *ConfigService
}

func NewConfigPresetService(repo *repository.ConfigPresetRepository, configService *ConfigService) *ConfigPresetService {
	return &ConfigPresetService{
		repository:    repo,
		configService: configService,
	}
}

func (s *ConfigPresetService) ListPresets(ctx context.Context, filter *model.ConfigPresetFilter) (*[]model.ConfigPreset, error) {
	return s.repository.GetAll(ctx, filter)
}

func (s *ConfigPresetService) GetPreset(ctx context.Context, id uuid.UUID) (*model.ConfigPreset, error) {
	preset, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if preset == nil {
		return nil, fmt.Errorf("preset not found")
	}
	return preset, nil
}

func (s *ConfigPresetService) CreatePreset(ctx context.Context, req *model.ConfigPresetCreateRequest) (*model.ConfigPreset, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("preset name is required")
	}
	if len(req.Name) > presetMaxNameLen {
		return nil, fmt.Errorf("preset name exceeds maximum length of %d characters", presetMaxNameLen)
	}
	if len(req.Description) > presetMaxDescLen {
		return nil, fmt.Errorf("preset description exceeds maximum length of %d characters", presetMaxDescLen)
	}

	preset := &model.ConfigPreset{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := marshalSections(req.Configuration, req.AssistRules, req.Event, req.EventRules, req.Settings, preset); err != nil {
		return nil, err
	}

	if err := s.repository.Insert(ctx, preset); err != nil {
		return nil, err
	}
	return preset, nil
}

func (s *ConfigPresetService) UpdatePreset(ctx context.Context, id uuid.UUID, req *model.ConfigPresetUpdateRequest) (*model.ConfigPreset, error) {
	preset, err := s.GetPreset(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		if len(*req.Name) == 0 {
			return nil, fmt.Errorf("preset name cannot be empty")
		}
		if len(*req.Name) > presetMaxNameLen {
			return nil, fmt.Errorf("preset name exceeds maximum length of %d characters", presetMaxNameLen)
		}
		preset.Name = *req.Name
	}
	if req.Description != nil {
		if len(*req.Description) > presetMaxDescLen {
			return nil, fmt.Errorf("preset description exceeds maximum length of %d characters", presetMaxDescLen)
		}
		preset.Description = *req.Description
	}
	if err := marshalSections(req.Configuration, req.AssistRules, req.Event, req.EventRules, req.Settings, preset); err != nil {
		return nil, err
	}

	if err := s.repository.Update(ctx, preset); err != nil {
		return nil, err
	}
	return preset, nil
}

func (s *ConfigPresetService) DeletePreset(ctx context.Context, id uuid.UUID) error {
	return s.repository.Delete(ctx, id)
}

func (s *ConfigPresetService) ApplyPreset(ctx context.Context, serverID string, presetID uuid.UUID, actorID, actorUsername string) ([]*model.Config, error) {
	preset, err := s.GetPreset(ctx, presetID)
	if err != nil {
		return nil, err
	}

	type sectionEntry struct {
		fileName string
		raw      string
	}

	sections := []sectionEntry{
		{ConfigurationJson, preset.Configuration},
		{AssistRulesJson, preset.AssistRules},
		{EventJson, preset.Event},
		{EventRulesJson, preset.EventRules},
		{SettingsJson, preset.Settings},
	}

	var results []*model.Config
	for _, section := range sections {
		if section.raw == "" {
			continue
		}
		var body map[string]interface{}
		if err := json.Unmarshal([]byte(section.raw), &body); err != nil {
			return nil, fmt.Errorf("failed to unmarshal preset section %s: %v", section.fileName, err)
		}
		result, err := s.configService.ApplySection(ctx, serverID, section.fileName, &body, actorID, actorUsername)
		if err != nil {
			return nil, fmt.Errorf("failed to apply preset section %s: %v", section.fileName, err)
		}
		if result != nil {
			results = append(results, result)
		}
	}

	return results, nil
}

// marshalSections marshals each non-nil config section into the preset's JSON string fields,
// enforcing the per-section byte size limit.
func marshalSections(configuration, assistRules, event, eventRules, settings *map[string]interface{}, preset *model.ConfigPreset) error {
	type entry struct {
		src  *map[string]interface{}
		name string
		dest *string
	}
	entries := []entry{
		{configuration, "configuration", &preset.Configuration},
		{assistRules, "assistRules", &preset.AssistRules},
		{event, "event", &preset.Event},
		{eventRules, "eventRules", &preset.EventRules},
		{settings, "settings", &preset.Settings},
	}
	for _, e := range entries {
		if e.src == nil {
			continue
		}
		b, err := json.Marshal(e.src)
		if err != nil {
			return fmt.Errorf("failed to marshal %s: %v", e.name, err)
		}
		if len(b) > presetMaxSectionBytes {
			return fmt.Errorf("%s section exceeds maximum size of %d bytes", e.name, presetMaxSectionBytes)
		}
		*e.dest = string(b)
	}
	return nil
}
