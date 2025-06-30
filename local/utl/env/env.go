package env

import (
	"os"
	"path/filepath"
)

const (
	// Default paths for when environment variables are not set
	DefaultSteamCMDPath = "c:\\steamcmd\\steamcmd.exe"
	DefaultNSSMPath     = ".\\nssm.exe"
)

// GetSteamCMDPath returns the SteamCMD executable path from environment variable or default
func GetSteamCMDPath() string {
	if path := os.Getenv("STEAMCMD_PATH"); path != "" {
		return path
	}
	return DefaultSteamCMDPath
}

// GetSteamCMDDirPath returns the directory containing SteamCMD executable
func GetSteamCMDDirPath() string {
	steamCMDPath := GetSteamCMDPath()
	return filepath.Dir(steamCMDPath)
}

// GetNSSMPath returns the NSSM executable path from environment variable or default
func GetNSSMPath() string {
	if path := os.Getenv("NSSM_PATH"); path != "" {
		return path
	}
	return DefaultNSSMPath
}

// ValidatePaths checks if the configured paths exist (optional validation)
func ValidatePaths() map[string]error {
	errors := make(map[string]error)

	// Check SteamCMD path
	steamCMDPath := GetSteamCMDPath()
	if _, err := os.Stat(steamCMDPath); os.IsNotExist(err) {
		errors["STEAMCMD_PATH"] = err
	}

	// Check NSSM path
	nssmPath := GetNSSMPath()
	if _, err := os.Stat(nssmPath); os.IsNotExist(err) {
		errors["NSSM_PATH"] = err
	}

	return errors
}
