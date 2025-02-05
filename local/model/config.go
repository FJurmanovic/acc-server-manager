package model

import "time"

// Config tracks configuration modifications
type Config struct {
	ID         uint      `gorm:"primaryKey"`
	ServerID   uint      `gorm:"not null"`
	ConfigFile string    `gorm:"not null"` // e.g. "settings.json"
	OldConfig  string    `gorm:"type:text"`
	NewConfig  string    `gorm:"type:text"`
	ChangedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

type Configurations struct {
	Configuration map[string]interface{}
	Entrylist     map[string]interface{}
	Event         map[string]interface{}
	EventRules    map[string]interface{}
	Settings      map[string]interface{}
}

type ServerSettings struct {
    ServerName                  string  `json:"serverName"`
    AdminPassword               string  `json:"adminPassword"`
    CarGroup                    string  `json:"carGroup"`
    TrackMedalsRequirement      int     `json:"trackMedalsRequirement"`
    SafetyRatingRequirement     int     `json:"safetyRatingRequirement"`
    RacecraftRatingRequirement  int     `json:"racecraftRatingRequirement"`
    Password                    string  `json:"password"`
    SpectatorPassword           string  `json:"spectatorPassword"`
    MaxCarSlots                 int     `json:"maxCarSlots"`
    DumpLeaderboards            int     `json:"dumpLeaderboards"`
    IsRaceLocked                int     `json:"isRaceLocked"`
    RandomizeTrackWhenEmpty     int     `json:"randomizeTrackWhenEmpty"`
    CentralEntryListPath        string  `json:"centralEntryListPath"`
    AllowAutoDQ                 int     `json:"allowAutoDQ"`
    ShortFormationLap           int     `json:"shortFormationLap"`
    FormationLapType            int     `json:"formationLapType"`
    IgnorePrematureDisconnects  int     `json:"ignorePrematureDisconnects"`
}

type EventConfig struct {
    Track                       string  `json:"track"`
    PreRaceWaitingTimeSeconds   int     `json:"preRaceWaitingTimeSeconds"`
    SessionOverTimeSeconds      int     `json:"sessionOverTimeSeconds"`
    AmbientTemp                 int     `json:"ambientTemp"`
    CloudLevel                  float64 `json:"cloudLevel"`
    Rain                        float64 `json:"rain"`
    WeatherRandomness           int     `json:"weatherRandomness"`
    PostQualySeconds            int     `json:"postQualySeconds"`
    PostRaceSeconds             int     `json:"postRaceSeconds"`
    SimracerWeatherConditions   int     `json:"simracerWeatherConditions"`
    IsFixedConditionQualification int   `json:"isFixedConditionQualification"`
    
    Sessions                    []Session `json:"sessions"`
}

type Session struct {
    HourOfDay          int     `json:"hourOfDay"`
    DayOfWeekend       int     `json:"dayOfWeekend"`
    TimeMultiplier     int     `json:"timeMultiplier"`
    SessionType        string  `json:"sessionType"`
    SessionDurationMinutes int `json:"sessionDurationMinutes"`
}

type EventRules struct {
    QualifyStandingType                int  `json:"qualifyStandingType"`
    PitWindowLengthSec                 int  `json:"pitWindowLengthSec"`
    DriverStintTimeSec                 int  `json:"driverStintTimeSec"`
    MandatoryPitstopCount              int  `json:"mandatoryPitstopCount"`
    MaxTotalDrivingTime                int  `json:"maxTotalDrivingTime"`
    IsRefuellingAllowedInRace          bool `json:"isRefuellingAllowedInRace"`
    IsRefuellingTimeFixed              bool `json:"isRefuellingTimeFixed"`
    IsMandatoryPitstopRefuellingRequired bool `json:"isMandatoryPitstopRefuellingRequired"`
    IsMandatoryPitstopTyreChangeRequired bool `json:"isMandatoryPitstopTyreChangeRequired"`
    IsMandatoryPitstopSwapDriverRequired bool `json:"isMandatoryPitstopSwapDriverRequired"`
    TyreSetCount                       int  `json:"tyreSetCount"`
}