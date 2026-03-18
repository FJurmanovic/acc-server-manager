package main

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/utl/configs"
	"flag"
	"fmt"
	"os"
)

func main() {
	var (
		encrypt   = flag.Bool("encrypt", false, "Encrypt a password")
		decrypt   = flag.Bool("decrypt", false, "Decrypt a password")
		password  = flag.String("password", "", "Password to encrypt/decrypt")
		help      = flag.Bool("help", false, "Show help")
	)

	flag.Parse()

	if *help || (!*encrypt && !*decrypt) {
		showHelp()
		return
	}

	if *encrypt && *decrypt {
		fmt.Fprintf(os.Stderr, "Error: Cannot specify both -encrypt and -decrypt\n")
		os.Exit(1)
	}

	if *password == "" {
		fmt.Fprintf(os.Stderr, "Error: Password is required\n")
		showHelp()
		os.Exit(1)
	}

	// Initialize configs to load encryption key
	configs.Init()

	if *encrypt {
		encrypted, err := model.EncryptPassword(*password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encrypting password: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(encrypted)
	}

	if *decrypt {
		decrypted, err := model.DecryptPassword(*password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error decrypting password: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(decrypted)
	}
}

func showHelp() {
	fmt.Println("Steam Credentials Encryption/Decryption Utility")
	fmt.Println()
	fmt.Println("This utility encrypts and decrypts Steam credentials using the same")
	fmt.Println("AES-256-GCM encryption used by the ACC Server Manager application.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  steam-crypt -encrypt -password \"your_password\"")
	fmt.Println("  steam-crypt -decrypt -password \"encrypted_string\"")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -encrypt    Encrypt the provided password")
	fmt.Println("  -decrypt    Decrypt the provided encrypted string")
	fmt.Println("  -password   The password to encrypt or encrypted string to decrypt")
	fmt.Println("  -help       Show this help message")
	fmt.Println()
	fmt.Println("Environment Variables Required:")
	fmt.Println("  ENCRYPTION_KEY  - 32-byte encryption key (same as main application)")
	fmt.Println("  APP_SECRET      - Application secret (required by configs)")
	fmt.Println("  APP_SECRET_CODE - Application secret code (required by configs)")
	fmt.Println("  ACCESS_KEY      - Access key (required by configs)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Encrypt a password")
	fmt.Println("  steam-crypt -encrypt -password \"mysteampassword\"")
	fmt.Println()
	fmt.Println("  # Decrypt an encrypted password")
	fmt.Println("  steam-crypt -decrypt -password \"base64encryptedstring\"")
	fmt.Println()
	fmt.Println("Security Notes:")
	fmt.Println("  - The encryption key must be exactly 32 bytes for AES-256")
	fmt.Println("  - Uses AES-256-GCM for authenticated encryption")
	fmt.Println("  - Each encryption includes a unique nonce for security")
	fmt.Println("  - Passwords are validated for length and basic security")
}