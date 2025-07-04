# ACC Server Manager Deployment Scripts

This directory contains scripts and tools for deploying the ACC Server Manager to a Windows server.

## Overview

The deployment process automates the following steps:
1. Build the application binaries for Windows
2. Copy binaries to the target server
3. Stop the Windows service
4. Backup current deployment
5. Deploy new binaries
6. Run database migrations
7. Start the Windows service
8. Verify deployment success

## Files

- `deploy.ps1` - Main PowerShell deployment script
- `deploy.bat` - Batch file wrapper for easier execution
- `deploy.config.example` - Configuration template
- `README.md` - This documentation

## Prerequisites

### Local Machine (Build Environment)
- Go 1.23 or later
- PowerShell (Windows PowerShell 5.1+ or PowerShell Core 7+)
- SSH client (OpenSSH recommended)
- SCP utility

### Target Windows Server
- Windows Server 2016 or later
- PowerShell 5.1 or later
- SSH Server (OpenSSH Server or similar)
- ACC Server Manager Windows service configured

## Setup

### 1. Configure SSH Access

Ensure SSH access to your Windows server:

```powershell
# On Windows Server, install OpenSSH Server
Add-WindowsCapability -Online -Name OpenSSH.Server~~~~0.0.1.0
Start-Service sshd
Set-Service -Name sshd -StartupType 'Automatic'
```

### 2. Create Deployment Configuration

Copy the configuration template and customize it:

```bash
cp deploy.config.example deploy.config
# Edit deploy.config with your server details
```

### 3. Set Up Windows Service

Ensure the ACC Server Manager is installed as a Windows service on the target server. You can use tools like NSSM or create a native Windows service.

## Usage

### Option 1: Using Batch Script (Recommended)

```batch
# Basic deployment
deploy.bat 192.168.1.100 admin "C:\AccServerManager"

# With additional options
deploy.bat 192.168.1.100 admin "C:\AccServerManager" -SkipTests

# Test deployment (dry run)
deploy.bat 192.168.1.100 admin "C:\AccServerManager" -WhatIf
```

### Option 2: Using PowerShell Script Directly

```powershell
# Basic deployment
.\deploy.ps1 -ServerHost "192.168.1.100" -ServerUser "admin" -DeployPath "C:\AccServerManager"

# With custom service name
.\deploy.ps1 -ServerHost "192.168.1.100" -ServerUser "admin" -DeployPath "C:\AccServerManager" -ServiceName "MyAccService"

# Skip tests and build
.\deploy.ps1 -ServerHost "192.168.1.100" -ServerUser "admin" -DeployPath "C:\AccServerManager" -SkipTests -SkipBuild

# Dry run
.\deploy.ps1 -ServerHost "192.168.1.100" -ServerUser "admin" -DeployPath "C:\AccServerManager" -WhatIf
```

### Option 3: Using GitHub Actions

The CI/CD pipeline automatically deploys when changes are pushed to the main branch. Configure the following secrets in your GitHub repository:

- `WINDOWS_SERVER_HOST` - Server hostname or IP
- `WINDOWS_SERVER_USER` - SSH username
- `WINDOWS_SERVER_SSH_KEY` - SSH private key
- `DEPLOY_PATH` - Deployment directory on server
- `SLACK_WEBHOOK_URL` - (Optional) Slack webhook for notifications

## Parameters

### Required Parameters
- `ServerHost` - Target Windows server hostname or IP address
- `ServerUser` - Username for SSH connection
- `DeployPath` - Full path where the application will be deployed

### Optional Parameters
- `ServiceName` - Windows service name (default: "ACC Server Manager")
- `BinaryName` - Main binary name (default: "acc-server-manager")
- `MigrateBinaryName` - Migration binary name (default: "acc-migrate")
- `SkipBuild` - Skip the build process
- `SkipTests` - Skip running tests
- `WhatIf` - Show what would be done without executing

## Deployment Process Details

### 1. Build Phase
- Runs Go tests (unless `-SkipTests` is specified)
- Builds API binary for Windows (GOOS=windows, GOARCH=amd64)
- Builds migration tool for Windows
- Creates binaries in `.\build` directory

### 2. Pre-deployment
- Copies binaries to server's temporary directory
- Creates deployment script on server

### 3. Deployment
- Stops the Windows service
- Creates backup of current deployment
- Copies new binaries to deployment directory
- Runs database migrations
- Starts the Windows service
- Verifies service is running

### 4. Rollback (if deployment fails)
- Restores from backup
- Restarts service with previous version
- Reports failure

### 5. Cleanup
- Removes temporary files
- Keeps last 5 backups (configurable)

## Troubleshooting

### Common Issues

#### SSH Connection Failed
```
Test-NetConnection -ComputerName <server> -Port 22
```
Ensure SSH server is running and accessible.

#### Service Not Found
Verify the service name matches exactly:
```powershell
Get-Service -Name "ACC Server Manager"
```

#### Permission Denied
Ensure the SSH user has permissions to:
- Stop/start the service
- Write to the deployment directory
- Execute PowerShell scripts

#### Build Failures
Check Go version and dependencies:
```bash
go version
go mod tidy
go test ./...
```

### Debug Mode

Run with verbose output:
```powershell
$VerbosePreference = "Continue"
.\deploy.ps1 -ServerHost "..." -ServerUser "..." -DeployPath "..." -Verbose
```

### Log Files

Deployment logs are stored in:
- Local: PowerShell transcript files
- Remote: Windows Event Log (Application log)

## Security Considerations

1. **SSH Keys**: Use SSH key authentication instead of passwords
2. **Service Account**: Run the service with minimal required permissions
3. **Firewall**: Restrict SSH access to deployment machines only
4. **Backup Encryption**: Consider encrypting backup files
5. **Secrets Management**: Use secure storage for configuration files

## Customization

### Custom Migration Scripts

Place custom migration scripts in the `scripts/migrations` directory. They will be executed in alphabetical order.

### Pre/Post Deployment Hooks

Configure custom scripts in the deployment configuration:
```ini
PreDeployScript=scripts/pre-deploy.ps1
PostDeployScript=scripts/post-deploy.ps1
```

### Environment Variables

Set environment variables during deployment:
```ini
EnvironmentVariables=ENV=production,LOG_LEVEL=info
```

## Monitoring and Notifications

### Slack Notifications
Configure Slack webhook URL to receive deployment notifications:
```ini
SlackWebhookUrl=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

### Health Checks
Configure health check endpoint to verify deployment:
```ini
HealthCheckUrl=http://localhost:8080/health
HealthCheckTimeout=60
```

## Best Practices

1. **Test deployments** in a staging environment first
2. **Use the `-WhatIf` flag** to preview changes
3. **Monitor service logs** after deployment
4. **Keep backups** of working deployments
5. **Use version tagging** for releases
6. **Document configuration changes**
7. **Test rollback procedures** regularly

## Support

For issues with deployment scripts:
1. Check the troubleshooting section
2. Review deployment logs
3. Verify server prerequisites
4. Test SSH connectivity manually
5. Check Windows service configuration

For application-specific issues, refer to the main project documentation.