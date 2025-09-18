package env

import (
	"os"
	"path/filepath"
)

const (
	DefaultSteamCMDPath = "c:\\steamcmd\\steamcmd.exe"
	DefaultNSSMPath     = ".\\nssm.exe"
)

func GetSteamCMDPath() string {
	if path := os.Getenv("STEAMCMD_PATH"); path != "" {
		return path
	}
	return DefaultSteamCMDPath
}

func GetSteamCMDDirPath() string {
	steamCMDPath := GetSteamCMDPath()
	return filepath.Dir(steamCMDPath)
}

func GetNSSMPath() string {
	if path := os.Getenv("NSSM_PATH"); path != "" {
		return path
	}
	return DefaultNSSMPath
}

func ValidatePaths() map[string]error {
	errors := make(map[string]error)

	steamCMDPath := GetSteamCMDPath()
	if _, err := os.Stat(steamCMDPath); os.IsNotExist(err) {
		errors["STEAMCMD_PATH"] = err
	}

	nssmPath := GetNSSMPath()
	if _, err := os.Stat(nssmPath); os.IsNotExist(err) {
		errors["NSSM_PATH"] = err
	}

	return errors
}
