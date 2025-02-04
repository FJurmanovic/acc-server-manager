package model

type ApiModel struct {
	Api string `json:"api"`
}

// ServiceStatus represents a Windows service state
type ServiceStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"` // "running", "stopped", "pending"
}
