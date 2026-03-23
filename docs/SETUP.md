# Setup Guide

This guide covers the complete setup process for ACC Server Manager.

## Prerequisites

### Required Software

1. **Windows OS**: Windows 10/11 or Windows Server 2016+
2. **Go**: Version 1.23.0+ ([Download](https://golang.org/dl/))
3. **SteamCMD**: For ACC server installation ([Download](https://steamcdn-a.akamaihd.net/client/installer/steamcmd.zip))
4. **NSSM**: For Windows service management ([Download](https://nssm.cc/release/nssm-2.24.zip))

### System Requirements

- Administrator privileges (for service and firewall management)
- At least 4GB RAM
- 10GB+ free disk space for ACC servers

## Installation Steps

### 1. Install SteamCMD

1. Download SteamCMD from the link above
2. Extract to `C:\steamcmd\`
3. Run `steamcmd.exe` once to complete setup

### 2. Install NSSM

1. Download NSSM from the link above
2. Extract the appropriate version (32-bit or 64-bit)
3. Copy `nssm.exe` to the ACC Server Manager directory or add to PATH

### 3. Build ACC Server Manager

```bash
# Clone the repository
git clone <repository-url>
cd acc-server-manager

# Build the application
go build -o api.exe cmd/api/main.go
```

### 4. Configure Environment

Run the setup script to generate secure configuration:

```powershell
# PowerShell
.\scripts\generate-secrets.ps1
```

This creates a `.env` file with secure random keys.

### 5. Set Tool Paths (Optional)

If your tools are not in the default locations:

```powershell
# PowerShell
$env:STEAMCMD_PATH = "D:\tools\steamcmd\steamcmd.exe"
$env:NSSM_PATH = "D:\tools\nssm\nssm.exe"
```

### 6. First Run

```bash
# Start the server
./api.exe
```

The server will start on http://localhost:3000

### 7. Initial Login

1. Open http://localhost:3000 in your browser
2. Default credentials:
   - Username: `admin`
   - Password: Set in `.env` file or use default (change immediately)

## Post-Installation

### Configure Steam Credentials

1. Log into the web interface
2. Go to Settings → Steam Configuration
3. Enter your Steam credentials (encrypted storage)

### Create Your First Server

1. Click "Add Server"
2. Enter server details
3. Click "Install" to download ACC server files via Steam
4. Configure and start your server

## Running as a Service

To run ACC Server Manager as a Windows service:

```bash
# Install service
nssm install "ACC Server Manager" "C:\path\to\api.exe"

# Start service
nssm start "ACC Server Manager"
```

## Verify Installation

Check that everything is working:

1. Access http://localhost:3000
2. Check logs in `logs/` directory
3. Try creating a test server

## Next Steps

- [Configure your servers](CONFIG.md)
- [Review API documentation](API.md)
- [Troubleshooting guide](TROUBLESHOOTING.md)