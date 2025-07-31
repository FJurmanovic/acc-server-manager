package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	Version       = "0.10.1"
	Prefix        = "v1"
	Secret        string
	SecretCode    string
	EncryptionKey string
	AccessKey     string
)

func Init() {
	godotenv.Load()
	// Fail fast if critical environment variables are missing
	Secret = getEnvRequired("APP_SECRET")
	SecretCode = getEnvRequired("APP_SECRET_CODE")
	EncryptionKey = getEnvRequired("ENCRYPTION_KEY")
	AccessKey = getEnvRequired("ACCESS_KEY")

	if len(EncryptionKey) != 32 {
		log.Fatal("ENCRYPTION_KEY must be exactly 32 bytes long for AES-256")
	}
}

// getEnv retrieves an environment variable or returns a fallback value.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Printf("Environment variable %s not set, using fallback.", key)
	return fallback
}

// getEnvRequired retrieves an environment variable and fails if it's not set.
// This should be used for critical configuration that must not have defaults.
func getEnvRequired(key string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	log.Fatalf("Required environment variable %s is not set or is empty", key)
	return "" // This line will never be reached due to log.Fatalf
}
