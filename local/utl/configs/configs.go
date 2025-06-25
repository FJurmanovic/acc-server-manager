package configs

import (
	"log"
	"os"
)

var (
	Version       = "0.0.1"
	Prefix        = "v1"
	Secret        string
	SecretCode    string
	EncryptionKey string
)

func init() {
	Secret = getEnv("APP_SECRET", "default-secret-for-dev-use-only")
	SecretCode = getEnv("APP_SECRET_CODE", "another-secret-for-dev-use-only")
	EncryptionKey = getEnv("ENCRYPTION_KEY", "a-secure-32-byte-long-key-!!!!!!") // Fallback MUST be 32 bytes for AES-256

	if len(EncryptionKey) != 32 {
		log.Fatal("ENCRYPTION_KEY must be 32 bytes long")
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
