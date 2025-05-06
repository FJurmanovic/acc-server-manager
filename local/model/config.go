package model

import "time"

// Config tracks configuration modifications
type Config struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	ServerID   uint      `json:"serverId" gorm:"not null"`
	ConfigFile string    `json:"configFile" gorm:"not null"` // e.g. "settings.json"
	OldConfig  string    `json:"oldConfig" gorm:"type:text"`
	NewConfig  string    `json:"newConfig" gorm:"type:text"`
	ChangedAt  time.Time `json:"changedAt" gorm:"default:CURRENT_TIMESTAMP"`
}

type Configurations struct {
	Configuration Configuration `json:"configuration"`
	AssistRules     AssistRules `json:"assistRules"`
	Event         EventConfig `json:"event"`
	EventRules    EventRules `json:"eventRules"`
	Settings      ServerSettings `json:"settings"`
}

type ServerSettings struct {
	ServerName                 string `json:"serverName"`
	AdminPassword              string `json:"adminPassword"`
	CarGroup                   string `json:"carGroup"`
	TrackMedalsRequirement     int    `json:"trackMedalsRequirement"`
	SafetyRatingRequirement    int    `json:"safetyRatingRequirement"`
	RacecraftRatingRequirement int    `json:"racecraftRatingRequirement"`
	Password                   string `json:"password"`
	SpectatorPassword          string `json:"spectatorPassword"`
	MaxCarSlots                int    `json:"maxCarSlots"`
	DumpLeaderboards           int    `json:"dumpLeaderboards"`
	IsRaceLocked               int    `json:"isRaceLocked"`
	RandomizeTrackWhenEmpty    int    `json:"randomizeTrackWhenEmpty"`
	CentralEntryListPath       string `json:"centralEntryListPath"`
	AllowAutoDQ                int    `json:"allowAutoDQ"`
	ShortFormationLap          int    `json:"shortFormationLap"`
	FormationLapType           int    `json:"formationLapType"`
	IgnorePrematureDisconnects int    `json:"ignorePrematureDisconnects"`
}

type EventConfig struct {
	Track                         string  `json:"track"`
	PreRaceWaitingTimeSeconds     int     `json:"preRaceWaitingTimeSeconds"`
	SessionOverTimeSeconds        int     `json:"sessionOverTimeSeconds"`
	AmbientTemp                   int     `json:"ambientTemp"`
	CloudLevel                    float64 `json:"cloudLevel"`
	Rain                          float64 `json:"rain"`
	WeatherRandomness             int     `json:"weatherRandomness"`
	PostQualySeconds              int     `json:"postQualySeconds"`
	PostRaceSeconds               int     `json:"postRaceSeconds"`
	SimracerWeatherConditions     int     `json:"simracerWeatherConditions"`
	IsFixedConditionQualification int     `json:"isFixedConditionQualification"`

	Sessions []Session `json:"sessions"`
}

type Session struct {
	HourOfDay              int    `json:"hourOfDay"`
	DayOfWeekend           int    `json:"dayOfWeekend"`
	TimeMultiplier         int    `json:"timeMultiplier"`
	SessionType            string `json:"sessionType"`
	SessionDurationMinutes int    `json:"sessionDurationMinutes"`
}

type AssistRules struct {
	StabilityControlLevelMax                  int  `json:"stabilityControlLevelMax"`
	DisableAutosteer                   int  `json:"disableAutosteer"`
	DisableAutoLights                   int  `json:"disableAutoLights"`
	DisableAutoWiper                int  `json:"disableAutoWiper"`
	DisableAutoEngineStart                  int  `json:"disableAutoEngineStart"`
	DisableAutoPitLimiter                         int  `json:"disableAutoPitLimiter"`
	DisableAutoGear                         int  `json:"disableAutoGear"`
	DisableAutoClutch                         int  `json:"disableAutoClutch"`
	DisableIdealLine                         int  `json:"disableIdealLine"`
}

type EventRules struct {
	QualifyStandingType                  int  `json:"qualifyStandingType"`
	PitWindowLengthSec                   int  `json:"pitWindowLengthSec"`
	DriverStintTimeSec                   int  `json:"driverStintTimeSec"`
	MandatoryPitstopCount                int  `json:"mandatoryPitstopCount"`
	MaxTotalDrivingTime                  int  `json:"maxTotalDrivingTime"`
	IsRefuellingAllowedInRace            bool `json:"isRefuellingAllowedInRace"`
	IsRefuellingTimeFixed                bool `json:"isRefuellingTimeFixed"`
	IsMandatoryPitstopRefuellingRequired bool `json:"isMandatoryPitstopRefuellingRequired"`
	IsMandatoryPitstopTyreChangeRequired bool `json:"isMandatoryPitstopTyreChangeRequired"`
	IsMandatoryPitstopSwapDriverRequired bool `json:"isMandatoryPitstopSwapDriverRequired"`
	TyreSetCount                         int  `json:"tyreSetCount"`
}

type Configuration struct {
	UdpPort              int    `json:"udpPort"`
	TcpPort           int    `json:"tcpPort"`
	MaxConnections         int    `json:"maxConnections"`
	LanDiscovery            int `json:"lanDiscovery"`
	RegisterToLobby int    `json:"registerToLobby"`
	ConfigVersion int    `json:"configVersion"`
}
