package model

type ServiceStatus string

const (
	StatusRunning    ServiceStatus = "SERVICE_RUNNING\r\n"
	StatusStopped    ServiceStatus = "SERVICE_STOPPED\r\n"
	StatusRestarting ServiceStatus = "SERVICE_RESTARTING\r\n"
)

type ApiModel struct {
	Api string `json:"api"`
}
