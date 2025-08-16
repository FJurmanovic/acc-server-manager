package security

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type PathValidator struct {
	allowedBasePaths []string
	blockedPatterns  []*regexp.Regexp
}

func NewPathValidator() *PathValidator {
	blockedPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\.\.`),
		regexp.MustCompile(`[<>:"|?*]`),
		regexp.MustCompile(`^(CON|PRN|AUX|NUL|COM[1-9]|LPT[1-9])$`),
		regexp.MustCompile(`\x00`),
		regexp.MustCompile(`^\\\\`),
		regexp.MustCompile(`^[a-zA-Z]:\\Windows`),
		regexp.MustCompile(`^[a-zA-Z]:\\Program Files`),
	}

	return &PathValidator{
		allowedBasePaths: []string{
			`C:\ACC-Servers`,
			`D:\ACC-Servers`,
			`E:\ACC-Servers`,
			`C:\SteamCMD`,
			`D:\SteamCMD`,
			`E:\SteamCMD`,
		},
		blockedPatterns: blockedPatterns,
	}
}

func (pv *PathValidator) ValidateInstallPath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	cleanPath := filepath.Clean(path)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("invalid path: %v", err)
	}

	for _, pattern := range pv.blockedPatterns {
		if pattern.MatchString(absPath) || pattern.MatchString(strings.ToUpper(filepath.Base(absPath))) {
			return fmt.Errorf("path contains forbidden patterns")
		}
	}

	allowed := false
	for _, basePath := range pv.allowedBasePaths {
		if strings.HasPrefix(strings.ToLower(absPath), strings.ToLower(basePath)) {
			allowed = true
			break
		}
	}

	if !allowed {
		return fmt.Errorf("path must be within allowed directories: %v", pv.allowedBasePaths)
	}

	if len(absPath) > 260 {
		return fmt.Errorf("path too long (max 260 characters)")
	}

	parentDir := filepath.Dir(absPath)
	if parentInfo, err := os.Stat(parentDir); err == nil {
		if !parentInfo.IsDir() {
			return fmt.Errorf("parent path is not a directory")
		}
	}

	return nil
}

func (pv *PathValidator) AddAllowedBasePath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid base path: %v", err)
	}

	pv.allowedBasePaths = append(pv.allowedBasePaths, absPath)
	return nil
}

func (pv *PathValidator) GetAllowedBasePaths() []string {
	return append([]string(nil), pv.allowedBasePaths...)
}