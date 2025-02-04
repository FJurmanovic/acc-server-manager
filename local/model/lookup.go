package model

// Track represents a track and its capacity
type Track struct {
	Name               string `json:"track" gorm:"primaryKey;size:50"`
	UniquePitBoxes     int    `json:"unique_pit_boxes"`
	PrivateServerSlots int    `json:"private_server_slots"`
}

// CarModel represents a car model mapping
type CarModel struct {
	Value    int    `json:"value" gorm:"primaryKey"`
	CarModel string `json:"car_model"`
}

// DriverCategory represents driver skill categories
type DriverCategory struct {
	Value    int    `json:"value" gorm:"primaryKey"`
	Category string `json:"category"`
}

// CupCategory represents championship cup categories
type CupCategory struct {
	Value    int    `json:"value" gorm:"primaryKey"`
	Category string `json:"category"`
}

// SessionType represents session types
type SessionType struct {
	Value       int    `json:"value" gorm:"primaryKey"`
	SessionType string `json:"session_type"`
}
