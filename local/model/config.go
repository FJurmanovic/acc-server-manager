package model

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

type IntString int

// Config tracks configuration modifications
type Config  struct {
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
	TrackMedalsRequirement     IntString    `json:"trackMedalsRequirement"`
	SafetyRatingRequirement    IntString    `json:"safetyRatingRequirement"`
	RacecraftRatingRequirement IntString    `json:"racecraftRatingRequirement"`
	Password                   string `json:"password"`
	SpectatorPassword          string `json:"spectatorPassword"`
	MaxCarSlots                IntString    `json:"maxCarSlots"`
	DumpLeaderboards           IntString    `json:"dumpLeaderboards"`
	IsRaceLocked               IntString    `json:"isRaceLocked"`
	RandomizeTrackWhenEmpty    IntString    `json:"randomizeTrackWhenEmpty"`
	CentralEntryListPath       string `json:"centralEntryListPath"`
	AllowAutoDQ                IntString    `json:"allowAutoDQ"`
	ShortFormationLap          IntString    `json:"shortFormationLap"`
	FormationLapType           IntString    `json:"formationLapType"`
	IgnorePrematureDisconnects IntString    `json:"ignorePrematureDisconnects"`
}

type EventConfig struct {
	Track                         string  `json:"track"`
	PreRaceWaitingTimeSeconds     IntString     `json:"preRaceWaitingTimeSeconds"`
	SessionOverTimeSeconds        IntString     `json:"sessionOverTimeSeconds"`
	AmbientTemp                   IntString     `json:"ambientTemp"`
	CloudLevel                    float64 `json:"cloudLevel"`
	Rain                          float64 `json:"rain"`
	WeatherRandomness             IntString     `json:"weatherRandomness"`
	PostQualySeconds              IntString     `json:"postQualySeconds"`
	PostRaceSeconds               IntString     `json:"postRaceSeconds"`
	SimracerWeatherConditions     IntString     `json:"simracerWeatherConditions"`
	IsFixedConditionQualification IntString     `json:"isFixedConditionQualification"`

	Sessions []Session `json:"sessions"`
}

type Session struct {
	HourOfDay              IntString    `json:"hourOfDay"`
	DayOfWeekend           IntString    `json:"dayOfWeekend"`
	TimeMultiplier         IntString    `json:"timeMultiplier"`
	SessionType            string `json:"sessionType"`
	SessionDurationMinutes IntString    `json:"sessionDurationMinutes"`
}

type AssistRules struct {
	StabilityControlLevelMax                  IntString  `json:"stabilityControlLevelMax"`
	DisableAutosteer                   IntString  `json:"disableAutosteer"`
	DisableAutoLights                   IntString  `json:"disableAutoLights"`
	DisableAutoWiper                IntString  `json:"disableAutoWiper"`
	DisableAutoEngineStart                  IntString  `json:"disableAutoEngineStart"`
	DisableAutoPitLimiter                         IntString  `json:"disableAutoPitLimiter"`
	DisableAutoGear                         IntString  `json:"disableAutoGear"`
	DisableAutoClutch                         IntString  `json:"disableAutoClutch"`
	DisableIdealLine                         IntString  `json:"disableIdealLine"`
}

type EventRules struct {
	QualifyStandingType                  IntString  `json:"qualifyStandingType"`
	PitWindowLengthSec                   IntString  `json:"pitWindowLengthSec"`
	DriverStIntStringTimeSec                   IntString  `json:"driverStIntStringTimeSec"`
	MandatoryPitstopCount                IntString  `json:"mandatoryPitstopCount"`
	MaxTotalDrivingTime                  IntString  `json:"maxTotalDrivingTime"`
	IsRefuellingAllowedInRace            bool `json:"isRefuellingAllowedInRace"`
	IsRefuellingTimeFixed                bool `json:"isRefuellingTimeFixed"`
	IsMandatoryPitstopRefuellingRequired bool `json:"isMandatoryPitstopRefuellingRequired"`
	IsMandatoryPitstopTyreChangeRequired bool `json:"isMandatoryPitstopTyreChangeRequired"`
	IsMandatoryPitstopSwapDriverRequired bool `json:"isMandatoryPitstopSwapDriverRequired"`
	TyreSetCount                         IntString  `json:"tyreSetCount"`
}

type Configuration struct {
	UdpPort              IntString    `json:"udpPort"`
	TcpPort           IntString    `json:"tcpPort"`
	MaxConnections         IntString    `json:"maxConnections"`
	LanDiscovery            IntString `json:"lanDiscovery"`
	RegisterToLobby IntString    `json:"registerToLobby"`
	ConfigVersion IntString    `json:"configVersion"`
}

type SystemConfig struct {
	ID            uint   `json:"id"`
	Key           string `json:"key"`
	Value         string `json:"value"`
	DefaultValue  string `json:"defaultValue"`
	Description   string `json:"description"`
	DateModified  string `json:"dateModified"`
}

// Known configuration keys
const (
	ConfigKeySteamCMDPath = "steamcmd_path"
	ConfigKeyNSSMPath     = "nssm_path"
)

// Cache keys
const (
	CacheKeySystemConfig = "system_config_%s"  // Format with config key
)

func (i *IntString) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err == nil {
		if (str == "") {
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

	return fmt.Errorf("invalid postQualySeconds value")
}

func (i IntString) ToString() string {
	return strconv.Itoa(int(i))
}

func (i IntString) ToInt() (int) {
	return int(i)
}

func (c *SystemConfig) Validate() error {
	if c.Key == "" {
		return fmt.Errorf("key is required")
	}

	// Validate paths exist for certain config keys
	switch c.Key {
	case ConfigKeySteamCMDPath, ConfigKeyNSSMPath:
		if c.Value == "" {
			if c.DefaultValue == "" {
				return fmt.Errorf("value or default value is required for path configuration")
			}
			// Use default value if value is empty
			c.Value = c.DefaultValue
		}
		
		// Check if path exists
		if _, err := os.Stat(c.Value); os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", c.Value)
		}
	}

	return nil
}

func (c *SystemConfig) GetEffectiveValue() string {
	if c.Value != "" {
		return c.Value
	}
	return c.DefaultValue
}