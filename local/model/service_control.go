package model

import (
	"database/sql/driver"
	"fmt"
	"strconv"
)

type ServiceStatus int

const (
	StatusUnknown ServiceStatus = iota
	StatusStopped
	StatusStopping
	StatusRestarting
	StatusStarting
	StatusRunning
)

// String converts the ServiceStatus to its string representation
func (s ServiceStatus) String() string {
	switch s {
	case StatusRunning:
		return "SERVICE_RUNNING"
	case StatusStopped:
		return "SERVICE_STOPPED"
	case StatusStarting:
		return "SERVICE_STARTING"
	case StatusStopping:
		return "SERVICE_STOPPING"
	case StatusRestarting:
		return "SERVICE_RESTARTING"
	default:
		return "SERVICE_UNKNOWN"
	}
}

// ParseServiceStatus converts a string to ServiceStatus
func ParseServiceStatus(s string) ServiceStatus {
	switch s {
	case "SERVICE_RUNNING":
		return StatusRunning
	case "SERVICE_STOPPED":
		return StatusStopped
	case "SERVICE_STARTING":
		return StatusStarting
	case "SERVICE_STOPPING":
		return StatusStopping
	case "SERVICE_RESTARTING":
		return StatusRestarting
	default:
		return StatusUnknown
	}
}

// MarshalJSON implements json.Marshaler interface
func (s ServiceStatus) MarshalJSON() ([]byte, error) {
	// Return the numeric value instead of string
	return []byte(strconv.Itoa(int(s))), nil
}

// UnmarshalJSON implements json.Unmarshaler interface
func (s *ServiceStatus) UnmarshalJSON(data []byte) error {
	// Try to parse as number first
	if i, err := strconv.Atoi(string(data)); err == nil {
		*s = ServiceStatus(i)
		return nil
	}

	// Fallback to string parsing for backward compatibility
	str := string(data)
	if len(str) >= 2 {
		// Remove quotes if present
		str = str[1 : len(str)-1]
	}
	*s = ParseServiceStatus(str)
	return nil
}

// Scan implements the sql.Scanner interface
func (s *ServiceStatus) Scan(value interface{}) error {
	if value == nil {
		*s = StatusUnknown
		return nil
	}

	switch v := value.(type) {
	case string:
		*s = ParseServiceStatus(v)
		return nil
	case []byte:
		*s = ParseServiceStatus(string(v))
		return nil
	case int64:
		*s = ServiceStatus(v)
		return nil
	default:
		return fmt.Errorf("unsupported type for ServiceStatus: %T", value)
	}
}

// Value implements the driver.Valuer interface
func (s ServiceStatus) Value() (driver.Value, error) {
	return s.String(), nil
}

type ServiceControlModel struct {
	ServiceControl string `json:"serviceControl"`
}
