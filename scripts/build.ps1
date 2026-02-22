param(
    [string]$BinaryName = "acc-server-manager",
    [string]$MigrateBinaryName = "acc-migrate",
    [string]$GOOS = "windows",
    [string]$GOARCH = "amd64",
    [string]$OutputPath = ".\build",
    [switch]$SkipTests
)

$ErrorActionPreference = "Stop"

function Write-Status {
    param([string]$Message, [string]$Color = "Green")
    Write-Host "[$(Get-Date -Format 'HH:mm:ss')] $Message" -ForegroundColor $Color
}

# Check if Go is installed
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw "Go is not installed or not in PATH"
}

# Clean output directory
if (Test-Path $OutputPath) {
    Remove-Item -Path $OutputPath -Recurse -Force
}
New-Item -ItemType Directory -Path $OutputPath -Force | Out-Null

# Run tests
if (-not $SkipTests) {
    Write-Status "Running tests..."
    & go test ./...
    if ($LASTEXITCODE -ne 0) {
        throw "Tests failed"
    }
}

# Set build environment
$env:GOOS   = $GOOS
$env:GOARCH = $GOARCH

$ext = if ($GOOS -eq "windows") { ".exe" } else { "" }

# Build API binary
Write-Status "Building API binary ($GOOS/$GOARCH)..."
& go build -o "$OutputPath\$BinaryName$ext" .\cmd\api
if ($LASTEXITCODE -ne 0) { throw "Failed to build API binary" }

# Build migration binary
Write-Status "Building migration binary ($GOOS/$GOARCH)..."
& go build -o "$OutputPath\$MigrateBinaryName$ext" .\cmd\migrate
if ($LASTEXITCODE -ne 0) { throw "Failed to build migration binary" }

# Reset environment
Remove-Item Env:\GOOS  -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue

Write-Status "Build completed -> $OutputPath"
