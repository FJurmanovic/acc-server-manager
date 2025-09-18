package model

type Track struct {
	Name               string `json:"track" gorm:"primaryKey;size:50"`
	UniquePitBoxes     int    `json:"unique_pit_boxes"`
	PrivateServerSlots int    `json:"private_server_slots"`
}

type CarModel struct {
	Value    int    `json:"value" gorm:"primaryKey"`
	CarModel string `json:"car_model"`
}

type DriverCategory struct {
	Value    int    `json:"value" gorm:"primaryKey"`
	Category string `json:"category"`
}

type CupCategory struct {
	Value    int    `json:"value" gorm:"primaryKey"`
	Category string `json:"category"`
}

type SessionType struct {
	Value       int    `json:"value" gorm:"primaryKey"`
	SessionType string `json:"session_type"`
}
