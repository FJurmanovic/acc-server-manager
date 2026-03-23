# Scripts

This directory contains utility scripts for ACC Server Manager.

## Setup Script

### generate-secrets.ps1 (Windows PowerShell)

Generates secure random secrets for the application and creates a `.env` file.

**Usage:**
```powershell
.\generate-secrets.ps1
```

This script:
- Generates a 64-byte JWT secret
- Generates 32-byte application secrets
- Generates a 32-character encryption key
- Creates a `.env` file with all required configuration

### generate-secrets.sh (Linux/macOS)

Same functionality for Unix-like systems.

**Usage:**
```bash
./generate-secrets.sh
```

## Manual Secret Generation

If you prefer to generate secrets manually:

```bash
# JWT Secret (64 bytes, base64 encoded)
openssl rand -base64 64

# Application secrets (32 bytes, hex encoded)
openssl rand -hex 32

# Encryption key (16 bytes = 32 hex characters)
openssl rand -hex 16
```

Then create `.env` file with:
```env
JWT_SECRET=your-jwt-secret
APP_SECRET=your-app-secret
APP_SECRET_CODE=your-app-secret-code
ENCRYPTION_KEY=your-32-char-hex-key
PORT=3000
DB_NAME=acc.db
```
