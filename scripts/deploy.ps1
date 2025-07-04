param(
    [Parameter(Mandatory=$true)]
    [string]$ServerHost,

    [Parameter(Mandatory=$true)]
    [string]$ServerUser,

    [Parameter(Mandatory=$true)]
    [string]$DeployPath,

    [string]$ServiceName = "ACC Server Manager",
    [string]$BinaryName = "acc-server-manager",
    [string]$MigrateBinaryName = "acc-migrate",
    [switch]$SkipBuild,
    [switch]$SkipTests,
    [switch]$WhatIf
)

# Configuration
$ErrorActionPreference = "Stop"
$LocalBuildPath = ".\build"
$RemoteTempPath = "C:\temp"

function Write-Status {
    param([string]$Message, [string]$Color = "Green")
    Write-Host "[$(Get-Date -Format 'HH:mm:ss')] $Message" -ForegroundColor $Color
}

function Write-Error-Status {
    param([string]$Message)
    Write-Host "[$(Get-Date -Format 'HH:mm:ss')] ERROR: $Message" -ForegroundColor Red
}

function Write-Warning-Status {
    param([string]$Message)
    Write-Host "[$(Get-Date -Format 'HH:mm:ss')] WARNING: $Message" -ForegroundColor Yellow
}

function Test-Prerequisites {
    Write-Status "Checking prerequisites..."

    # Check if Go is installed
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        throw "Go is not installed or not in PATH"
    }

    # Check if we can connect to the server
    $testConnection = Test-NetConnection -ComputerName $ServerHost -Port 22 -WarningAction SilentlyContinue
    if (-not $testConnection.TcpTestSucceeded) {
        Write-Warning-Status "Cannot connect to $ServerHost on port 22. Make sure SSH is enabled."
    }

    Write-Status "Prerequisites check completed"
}

function Build-Application {
    if ($SkipBuild) {
        Write-Status "Skipping build (SkipBuild flag set)"
        return
    }

    Write-Status "Building application..."

    # Clean build directory
    if (Test-Path $LocalBuildPath) {
        Remove-Item -Path $LocalBuildPath -Recurse -Force
    }
    New-Item -ItemType Directory -Path $LocalBuildPath -Force | Out-Null

    # Run tests
    if (-not $SkipTests) {
        Write-Status "Running tests..."
        & go test -v ./...
        if ($LASTEXITCODE -ne 0) {
            throw "Tests failed"
        }
    }

    # Build API for Windows
    Write-Status "Building API binary..."
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    & go build -o "$LocalBuildPath\$BinaryName.exe" .\cmd\api
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to build API binary"
    }

    # Build migration tool for Windows
    Write-Status "Building migration binary..."
    & go build -o "$LocalBuildPath\$MigrateBinaryName.exe" .\cmd\migrate
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to build migration binary"
    }

    # Reset environment variables
    Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue

    Write-Status "Build completed successfully"
}

function Copy-FilesToServer {
    Write-Status "Copying files to server..."

    if ($WhatIf) {
        Write-Status "WHAT-IF: Would copy files to $ServerHost" -Color "Cyan"
        return
    }

    # Copy binaries to server
    scp -o StrictHostKeyChecking=no "$LocalBuildPath\$BinaryName.exe" "${ServerUser}@${ServerHost}:${RemoteTempPath}/"
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to copy API binary to server"
    }

    scp -o StrictHostKeyChecking=no "$LocalBuildPath\$MigrateBinaryName.exe" "${ServerUser}@${ServerHost}:${RemoteTempPath}/"
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to copy migration binary to server"
    }

    Write-Status "Files copied successfully"
}

function Deploy-ToServer {
    Write-Status "Deploying to server..."

    if ($WhatIf) {
        Write-Status "WHAT-IF: Would deploy to $ServerHost" -Color "Cyan"
        return
    }

    # Create deployment script
    $deployScript = @"
Write-Host "Starting deployment process..." -ForegroundColor Green

# Check if service exists and stop it
`$service = Get-Service -Name '$ServiceName' -ErrorAction SilentlyContinue
if (`$service) {
    Write-Host "Stopping service: $ServiceName" -ForegroundColor Yellow
    Stop-Service -Name '$ServiceName' -Force

    # Wait for service to stop
    `$timeout = 30
    `$elapsed = 0
    while (`$service.Status -ne 'Stopped' -and `$elapsed -lt `$timeout) {
        Start-Sleep -Seconds 1
        `$elapsed++
        `$service.Refresh()
    }

    if (`$service.Status -ne 'Stopped') {
        Write-Error "Failed to stop service within timeout"
        exit 1
    }
    Write-Host "Service stopped successfully" -ForegroundColor Green
} else {
    Write-Host "Service not found: $ServiceName" -ForegroundColor Yellow
}

# Create backup of current deployment
`$backupPath = "$DeployPath\backup_`$(Get-Date -Format 'yyyyMMdd_HHmmss')"
if (Test-Path "$DeployPath\$BinaryName.exe") {
    Write-Host "Creating backup at: `$backupPath" -ForegroundColor Yellow
    New-Item -ItemType Directory -Path `$backupPath -Force | Out-Null
    Copy-Item "$DeployPath\*" -Destination `$backupPath -Recurse -Force
}

# Copy new binaries
Write-Host "Copying new binaries to: $DeployPath" -ForegroundColor Yellow
if (-not (Test-Path "$DeployPath")) {
    New-Item -ItemType Directory -Path "$DeployPath" -Force | Out-Null
}
Copy-Item "$RemoteTempPath\$BinaryName.exe" -Destination "$DeployPath\$BinaryName.exe" -Force
Copy-Item "$RemoteTempPath\$MigrateBinaryName.exe" -Destination "$DeployPath\$MigrateBinaryName.exe" -Force

# Run migrations
Write-Host "Running database migrations..." -ForegroundColor Yellow
try {
    `$migrateResult = & "$DeployPath\$MigrateBinaryName.exe" 2>&1
    Write-Host "Migration output: `$migrateResult" -ForegroundColor Cyan
} catch {
    Write-Warning "Migration failed: `$_"
}

# Start service
if (`$service) {
    Write-Host "Starting service: $ServiceName" -ForegroundColor Yellow
    Start-Service -Name '$ServiceName'

    # Wait for service to start
    `$timeout = 30
    `$elapsed = 0
    while (`$service.Status -ne 'Running' -and `$elapsed -lt `$timeout) {
        Start-Sleep -Seconds 1
        `$elapsed++
        `$service.Refresh()
    }

    if (`$service.Status -ne 'Running') {
        Write-Error "Failed to start service within timeout"
        # Rollback
        Write-Host "Rolling back deployment..." -ForegroundColor Red
        if (Test-Path `$backupPath) {
            Copy-Item "`$backupPath\*" -Destination "$DeployPath" -Recurse -Force
            Start-Service -Name '$ServiceName'
        }
        exit 1
    }
    Write-Host "Service started successfully" -ForegroundColor Green
} else {
    Write-Host "Service not configured. Manual start required." -ForegroundColor Yellow
}

# Cleanup old backups (keep last 5)
`$backupDir = Split-Path "$DeployPath" -Parent
Get-ChildItem -Path `$backupDir -Directory -Name "backup_*" |
    Sort-Object -Descending |
    Select-Object -Skip 5 |
    ForEach-Object { Remove-Item -Path "`$backupDir\`$_" -Recurse -Force }

# Cleanup temp files
Remove-Item "$RemoteTempPath\$BinaryName.exe" -Force -ErrorAction SilentlyContinue
Remove-Item "$RemoteTempPath\$MigrateBinaryName.exe" -Force -ErrorAction SilentlyContinue

Write-Host "Deployment completed successfully!" -ForegroundColor Green
"@

    # Save script to temp file
    $tempScript = [System.IO.Path]::GetTempFileName() + ".ps1"
    $deployScript | Out-File -FilePath $tempScript -Encoding UTF8

    try {
        # Copy deployment script to server
        scp -o StrictHostKeyChecking=no $tempScript "${ServerUser}@${ServerHost}:${RemoteTempPath}/deploy_script.ps1"
        if ($LASTEXITCODE -ne 0) {
            throw "Failed to copy deployment script to server"
        }

        # Execute deployment script on server
        ssh -o StrictHostKeyChecking=no "${ServerUser}@${ServerHost}" "powershell.exe -ExecutionPolicy Bypass -File ${RemoteTempPath}/deploy_script.ps1"
        if ($LASTEXITCODE -ne 0) {
            throw "Deployment script failed on server"
        }

        # Cleanup deployment script
        ssh -o StrictHostKeyChecking=no "${ServerUser}@${ServerHost}" "del ${RemoteTempPath}/deploy_script.ps1"

    } finally {
        # Remove temp script
        Remove-Item $tempScript -Force -ErrorAction SilentlyContinue
    }

    Write-Status "Deployment completed successfully"
}

function Show-Usage {
    Write-Host @"
ACC Server Manager Deployment Script

Usage: .\deploy.ps1 -ServerHost <host> -ServerUser <user> -DeployPath <path> [options]

Parameters:
  -ServerHost       Target Windows server hostname or IP
  -ServerUser       Username for SSH connection
  -DeployPath       Deployment directory on target server
  -ServiceName      Windows service name (default: "ACC Server Manager")
  -BinaryName       Main binary name (default: "acc-server-manager")
  -MigrateBinaryName Migration binary name (default: "acc-migrate")
  -SkipBuild        Skip the build process
  -SkipTests        Skip running tests
  -WhatIf           Show what would be done without executing

Examples:
  .\deploy.ps1 -ServerHost "192.168.1.100" -ServerUser "admin" -DeployPath "C:\AccServerManager"
  .\deploy.ps1 -ServerHost "server.example.com" -ServerUser "deploy" -DeployPath "C:\Services\AccServerManager" -SkipTests
  .\deploy.ps1 -ServerHost "192.168.1.100" -ServerUser "admin" -DeployPath "C:\AccServerManager" -WhatIf

Requirements:
  - Go installed locally
  - SSH client (OpenSSH)
  - SCP utility
  - PowerShell on target Windows server
  - SSH server running on target Windows server

"@ -ForegroundColor Cyan
}

# Main execution
try {
    Write-Status "Starting ACC Server Manager deployment" -Color "Cyan"
    Write-Status "Target: $ServerUser@$ServerHost"
    Write-Status "Deploy Path: $DeployPath"
    Write-Status "Service Name: $ServiceName"

    if ($WhatIf) {
        Write-Status "WHAT-IF MODE: No changes will be made" -Color "Cyan"
    }

    Test-Prerequisites
    Build-Application
    Copy-FilesToServer
    Deploy-ToServer

    Write-Status "Deployment completed successfully!" -Color "Green"

} catch {
    Write-Error-Status $_.Exception.Message
    Write-Host "Deployment failed. Check the error above for details." -ForegroundColor Red
    exit 1
}
