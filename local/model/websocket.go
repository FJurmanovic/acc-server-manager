package model

import (
	"github.com/google/uuid"
)

// ServerCreationStep represents the steps in server creation process
type ServerCreationStep string

const (
	StepValidation        ServerCreationStep = "validation"
	StepDirectoryCreation ServerCreationStep = "directory_creation"
	StepSteamDownload     ServerCreationStep = "steam_download"
	StepConfigGeneration  ServerCreationStep = "config_generation"
	StepServiceCreation   ServerCreationStep = "service_creation"
	StepFirewallRules     ServerCreationStep = "firewall_rules"
	StepDatabaseSave      ServerCreationStep = "database_save"
	StepCompleted         ServerCreationStep = "completed"
)

// StepStatus represents the status of a step
type StepStatus string

const (
	StatusPending    StepStatus = "pending"
	StatusInProgress StepStatus = "in_progress"
	StatusCompleted  StepStatus = "completed"
	StatusFailed     StepStatus = "failed"
)

// WebSocketMessageType represents different types of WebSocket messages
type WebSocketMessageType string

const (
	MessageTypeStep        WebSocketMessageType = "step"
	MessageTypeSteamOutput WebSocketMessageType = "steam_output"
	MessageTypeError       WebSocketMessageType = "error"
	MessageTypeComplete    WebSocketMessageType = "complete"
)

// WebSocketMessage is the base structure for all WebSocket messages
type WebSocketMessage struct {
	Type      WebSocketMessageType `json:"type"`
	ServerID  *uuid.UUID           `json:"server_id,omitempty"`
	Timestamp int64                `json:"timestamp"`
	Data      interface{}          `json:"data"`
}

// StepMessage represents a step update message
type StepMessage struct {
	Step    ServerCreationStep `json:"step"`
	Status  StepStatus         `json:"status"`
	Message string             `json:"message,omitempty"`
	Error   string             `json:"error,omitempty"`
}

// SteamOutputMessage represents SteamCMD output
type SteamOutputMessage struct {
	Output  string `json:"output"`
	IsError bool   `json:"is_error"`
}

// ErrorMessage represents an error message
type ErrorMessage struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// CompleteMessage represents completion message
type CompleteMessage struct {
	ServerID uuid.UUID `json:"server_id"`
	Success  bool      `json:"success"`
	Message  string    `json:"message"`
}

// GetStepDescription returns a human-readable description for each step
func GetStepDescription(step ServerCreationStep) string {
	descriptions := map[ServerCreationStep]string{
		StepValidation:        "Validating server configuration",
		StepDirectoryCreation: "Creating server directories",
		StepSteamDownload:     "Downloading server files via Steam",
		StepConfigGeneration:  "Generating server configuration files",
		StepServiceCreation:   "Creating Windows service",
		StepFirewallRules:     "Configuring firewall rules",
		StepDatabaseSave:      "Saving server to database",
		StepCompleted:         "Server creation completed",
	}
	return descriptions[step]
}
