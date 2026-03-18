package env

import (
	"os"
	"path/filepath"
)

const (
	DefaultSteamCMDPath      = "c:\\steamcmd\\steamcmd.exe"
	DefaultNSSMPath          = ".\\nssm.exe"
	DefaultPlatform          = "windows"
	DefaultACCImage          = "acc-wine:latest"
	DefaultACCServersPath    = "/data/servers"
	DefaultLinuxSteamCMDPath = "/usr/games/steamcmd"
)

func GetPlatform() string {
	if p := os.Getenv("PLATFORM"); p != "" {
		return p
	}
	return DefaultPlatform
}

func IsDockerPlatform() bool {
	return GetPlatform() == "docker"
}

func GetSteamCMDPath() string {
	if path := os.Getenv("STEAMCMD_PATH"); path != "" {
		return path
	}
	return DefaultSteamCMDPath
}

func GetLinuxSteamCMDPath() string {
	if path := os.Getenv("STEAMCMD_LINUX_PATH"); path != "" {
		return path
	}
	return DefaultLinuxSteamCMDPath
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

func GetACCImage() string {
	if img := os.Getenv("ACC_IMAGE"); img != "" {
		return img
	}
	return DefaultACCImage
}

func GetACCServersPath() string {
	if p := os.Getenv("ACC_SERVERS_PATH"); p != "" {
		return p
	}
	return DefaultACCServersPath
}

func ValidatePaths() map[string]error {
	errors := make(map[string]error)

	if !IsDockerPlatform() {
		steamCMDPath := GetSteamCMDPath()
		if _, err := os.Stat(steamCMDPath); os.IsNotExist(err) {
			errors["STEAMCMD_PATH"] = err
		}

		nssmPath := GetNSSMPath()
		if _, err := os.Stat(nssmPath); os.IsNotExist(err) {
			errors["NSSM_PATH"] = err
		}
	}

	return errors
}
