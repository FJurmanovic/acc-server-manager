package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IntString int
type IntBool int

// Config tracks configuration modifications
type Config struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	ServerID   uuid.UUID `json:"serverId" gorm:"not null;type:uuid"`
	ConfigFile string    `json:"configFile" gorm:"not null"` // e.g. "settings.json"
	OldConfig  string    `json:"oldConfig" gorm:"type:text"`
	NewConfig  string    `json:"newConfig" gorm:"type:text"`
	ChangedAt  time.Time `json:"changedAt" gorm:"default:CURRENT_TIMESTAMP"`
}

// BeforeCreate is a GORM hook that runs before creating new config entries
func (c *Config) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.ChangedAt.IsZero() {
		c.ChangedAt = time.Now().UTC()
	}
	return nil
}

type Configurations struct {
	Configuration Configuration  `json:"configuration"`
	AssistRules   AssistRules    `json:"assistRules"`
	Event         EventConfig    `json:"event"`
	EventRules    EventRules     `json:"eventRules"`
	Settings      ServerSettings `json:"settings"`
}

type ServerSettings struct {
	ServerName                 string    `json:"serverName"`
	AdminPassword              string    `json:"adminPassword"`
	CarGroup                   string    `json:"carGroup"`
	TrackMedalsRequirement     IntString `json:"trackMedalsRequirement"`
	SafetyRatingRequirement    IntString `json:"safetyRatingRequirement"`
	RacecraftRatingRequirement IntString `json:"racecraftRatingRequirement"`
	Password                   string    `json:"password"`
	SpectatorPassword          string    `json:"spectatorPassword"`
	MaxCarSlots                IntString `json:"maxCarSlots"`
	DumpLeaderboards           IntString `json:"dumpLeaderboards"`
	IsRaceLocked               IntString `json:"isRaceLocked"`
	RandomizeTrackWhenEmpty    IntString `json:"randomizeTrackWhenEmpty"`
	CentralEntryListPath       string    `json:"centralEntryListPath"`
	AllowAutoDQ                IntString `json:"allowAutoDQ"`
	ShortFormationLap          IntString `json:"shortFormationLap"`
	FormationLapType           IntString `json:"formationLapType"`
	IgnorePrematureDisconnects IntString `json:"ignorePrematureDisconnects"`
}

type EventConfig struct {
	Track                         string    `json:"track"`
	PreRaceWaitingTimeSeconds     IntString `json:"preRaceWaitingTimeSeconds"`
	SessionOverTimeSeconds        IntString `json:"sessionOverTimeSeconds"`
	AmbientTemp                   IntString `json:"ambientTemp"`
	CloudLevel                    float64   `json:"cloudLevel"`
	Rain                          float64   `json:"rain"`
	WeatherRandomness             IntString `json:"weatherRandomness"`
	PostQualySeconds              IntString `json:"postQualySeconds"`
	PostRaceSeconds               IntString `json:"postRaceSeconds"`
	SimracerWeatherConditions     IntString `json:"simracerWeatherConditions"`
	IsFixedConditionQualification IntString `json:"isFixedConditionQualification"`

	Sessions []Session `json:"sessions"`
}

type Session struct {
	HourOfDay              IntString    `json:"hourOfDay"`
	DayOfWeekend           IntString    `json:"dayOfWeekend"`
	TimeMultiplier         IntString    `json:"timeMultiplier"`
	SessionType            TrackSession `json:"sessionType"`
	SessionDurationMinutes IntString    `json:"sessionDurationMinutes"`
}

type AssistRules struct {
	StabilityControlLevelMax IntString `json:"stabilityControlLevelMax"`
	DisableAutosteer         IntString `json:"disableAutosteer"`
	DisableAutoLights        IntString `json:"disableAutoLights"`
	DisableAutoWiper         IntString `json:"disableAutoWiper"`
	DisableAutoEngineStart   IntString `json:"disableAutoEngineStart"`
	DisableAutoPitLimiter    IntString `json:"disableAutoPitLimiter"`
	DisableAutoGear          IntString `json:"disableAutoGear"`
	DisableAutoClutch        IntString `json:"disableAutoClutch"`
	DisableIdealLine         IntString `json:"disableIdealLine"`
}

type EventRules struct {
	QualifyStandingType                  IntString `json:"qualifyStandingType"`
	PitWindowLengthSec                   IntString `json:"pitWindowLengthSec"`
	DriverStIntStringTimeSec             IntString `json:"driverStIntStringTimeSec"`
	MandatoryPitstopCount                IntString `json:"mandatoryPitstopCount"`
	MaxTotalDrivingTime                  IntString `json:"maxTotalDrivingTime"`
	IsRefuellingAllowedInRace            IntBool   `json:"isRefuellingAllowedInRace"`
	IsRefuellingTimeFixed                IntBool   `json:"isRefuellingTimeFixed"`
	IsMandatoryPitstopRefuellingRequired IntBool   `json:"isMandatoryPitstopRefuellingRequired"`
	IsMandatoryPitstopTyreChangeRequired IntBool   `json:"isMandatoryPitstopTyreChangeRequired"`
	IsMandatoryPitstopSwapDriverRequired IntBool   `json:"isMandatoryPitstopSwapDriverRequired"`
	TyreSetCount                         IntString `json:"tyreSetCount"`
}

type Configuration struct {
	UdpPort         IntString `json:"udpPort"`
	TcpPort         IntString `json:"tcpPort"`
	MaxConnections  IntString `json:"maxConnections"`
	LanDiscovery    IntString `json:"lanDiscovery"`
	RegisterToLobby IntString `json:"registerToLobby"`
	ConfigVersion   IntString `json:"configVersion"`
}

// Known configuration keys

func (i *IntBool) UnmarshalJSON(b []byte) error {
	var str int
	if err := json.Unmarshal(b, &str); err == nil && str <= 1 {
		*i = IntBool(str)
		return nil
	}

	var num bool
	if err := json.Unmarshal(b, &num); err == nil {
		if num {
			*i = IntBool(1)
		} else {
			*i = IntBool(0)
		}
		return nil
	}

	return fmt.Errorf("invalid IntBool value")
}

func (i IntBool) ToInt() int {
	return int(i)
}

func (i IntBool) ToBool() bool {
	return i == 1
}

func (i *IntString) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err == nil {
		if str == "" {
			*i = IntString(0)
		} else {
			n, err := strconv.Atoi(str)
			if err != nil {
				return err
			}
			*i = IntString(n)
		}
		return nil
	}

	var num int
	if err := json.Unmarshal(b, &num); err == nil {
		*i = IntString(num)
		return nil
	}

	return fmt.Errorf("invalid IntString value")
}

func (i IntString) ToString() string {
	return strconv.Itoa(int(i))
}

func (i IntString) ToInt() int {
	return int(i)
}
