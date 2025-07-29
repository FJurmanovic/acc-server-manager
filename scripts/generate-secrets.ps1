# ACC Server Manager - Secret Generation Script
# This script generates cryptographically secure secrets for the ACC Server Manager

Write-Host "ACC Server Manager - Secret Generation Script" -ForegroundColor Green
Write-Host "=============================================" -ForegroundColor Green
Write-Host ""

# Function to generate random bytes and convert to hex
function Generate-HexString {
    param([int]$Length)
    $bytes = New-Object byte[] $Length
    $rng = [System.Security.Cryptography.RNGCryptoServiceProvider]::new()
    $rng.GetBytes($bytes)
    $rng.Dispose()
    return [System.BitConverter]::ToString($bytes) -replace '-', ''
}

# Function to generate random bytes and convert to base64
function Generate-Base64String {
    param([int]$Length)
    $bytes = New-Object byte[] $Length
    $rng = [System.Security.Cryptography.RNGCryptoServiceProvider]::new()
    $rng.GetBytes($bytes)
    $rng.Dispose()
    return [System.Convert]::ToBase64String($bytes)
}

# Generate secrets
Write-Host "Generating cryptographically secure secrets..." -ForegroundColor Yellow
Write-Host ""

$jwtSecret = Generate-Base64String -Length 64
$appSecret = Generate-HexString -Length 32
$appSecretCode = Generate-HexString -Length 32
$accessKey = Generate-HexString -Length 32
$encryptionKey = Generate-HexString -Length 16

# Display generated secrets
Write-Host "Generated Secrets:" -ForegroundColor Cyan
Write-Host "==================" -ForegroundColor Cyan
Write-Host ""
Write-Host "JWT_SECRET=" -NoNewline -ForegroundColor White
Write-Host $jwtSecret -ForegroundColor Yellow
Write-Host ""
Write-Host "APP_SECRET=" -NoNewline -ForegroundColor White
Write-Host $appSecret -ForegroundColor Yellow
Write-Host ""
Write-Host "APP_SECRET_CODE=" -NoNewline -ForegroundColor White
Write-Host $appSecretCode -ForegroundColor Yellow
Write-Host ""
Write-Host "ENCRYPTION_KEY=" -NoNewline -ForegroundColor White
Write-Host $encryptionKey -ForegroundColor Yellow
Write-Host ""
Write-Host "ACCESS_KEY=" -NoNewline -ForegroundColor White
Write-Host $accessKey -ForegroundColor Yellow
Write-Host ""

# Check if .env file exists
$envFile = ".env"
$envExists = Test-Path $envFile

if ($envExists) {
    Write-Host "Warning: .env file already exists!" -ForegroundColor Red
    $overwrite = Read-Host "Do you want to update it with new secrets? (y/N)"
    if ($overwrite -eq "y" -or $overwrite -eq "Y") {
        $updateFile = $true
    } else {
        $updateFile = $false
        Write-Host "Secrets generated but not written to file." -ForegroundColor Yellow
    }
} else {
    $createFile = Read-Host "Create .env file with these secrets? (Y/n)"
    if ($createFile -eq "" -or $createFile -eq "y" -or $createFile -eq "Y") {
        $updateFile = $true
    } else {
        $updateFile = $false
        Write-Host "Secrets generated but not written to file." -ForegroundColor Yellow
    }
}

if ($updateFile) {
    # Create or update .env file
    if ($envExists) {
        # Backup existing file
        $backupFile = ".env.backup." + (Get-Date -Format "yyyyMMdd-HHmmss")
        Copy-Item $envFile $backupFile
        Write-Host "Backed up existing .env to $backupFile" -ForegroundColor Green

        # Read existing content and update secrets
        $content = Get-Content $envFile
        $newContent = @()

        foreach ($line in $content) {
            if ($line -match "^JWT_SECRET=") {
                $newContent += "JWT_SECRET=$jwtSecret"
            } elseif ($line -match "^APP_SECRET=") {
                $newContent += "APP_SECRET=$appSecret"
            } elseif ($line -match "^APP_SECRET_CODE=") {
                $newContent += "APP_SECRET_CODE=$appSecretCode"
            } elseif ($line -match "^ENCRYPTION_KEY=") {
                $newContent += "ENCRYPTION_KEY=$encryptionKey"
            } elseif ($line -match "^ACCESS_KEY=") {
                $newContent += "ACCESS_KEY=$accessKey"
            } else {
                $newContent += $line
            }
        }

        $newContent | Out-File -FilePath $envFile -Encoding UTF8
        Write-Host "Updated .env file with new secrets" -ForegroundColor Green
    } else {
        # Create new .env file from template
        if (Test-Path ".env.example") {
            $template = Get-Content ".env.example"
            $newContent = @()

            foreach ($line in $template) {
                if ($line -match "^JWT_SECRET=") {
                    $newContent += "JWT_SECRET=$jwtSecret"
                } elseif ($line -match "^APP_SECRET=") {
                    $newContent += "APP_SECRET=$appSecret"
                } elseif ($line -match "^APP_SECRET_CODE=") {
                    $newContent += "APP_SECRET_CODE=$appSecretCode"
                } elseif ($line -match "^ENCRYPTION_KEY=") {
                    $newContent += "ENCRYPTION_KEY=$encryptionKey"
                } elseif ($line -match "^ACCESS_KEY=") {
                    $newContent += "ACCESS_KEY=$accessKey"
                } else {
                    $newContent += $line
                }
            }

            $newContent | Out-File -FilePath $envFile -Encoding UTF8
            Write-Host "Created .env file from template with generated secrets" -ForegroundColor Green
        } else {
            # Create minimal .env file
            $minimalEnv = @(
                "# ACC Server Manager Environment Configuration",
                "# Generated on $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')",
                "",
                "# CRITICAL SECURITY SETTINGS (REQUIRED)",
                "JWT_SECRET=$jwtSecret",
                "APP_SECRET=$appSecret",
                "APP_SECRET_CODE=$appSecretCode",
                "ENCRYPTION_KEY=$encryptionKey",
                "ACCESS_KEY=$accessKey",
                "",
                "# CORE APPLICATION SETTINGS",
                "DB_NAME=acc.db",
                "PORT=3000",
                "CORS_ALLOWED_ORIGIN=http://localhost:5173",
                "PASSWORD=change-this-default-admin-password"
            )

            $minimalEnv | Out-File -FilePath $envFile -Encoding UTF8
            Write-Host "Created minimal .env file with generated secrets" -ForegroundColor Green
        }
    }
}

Write-Host ""
Write-Host "Security Notes:" -ForegroundColor Red
Write-Host "===============" -ForegroundColor Red
Write-Host "1. Keep these secrets secure and never commit them to version control" -ForegroundColor Yellow
Write-Host "2. Use different secrets for each environment (dev, staging, production)" -ForegroundColor Yellow
Write-Host "3. Rotate secrets regularly in production environments" -ForegroundColor Yellow
Write-Host "4. The ENCRYPTION_KEY is exactly 32 bytes as required for AES-256" -ForegroundColor Yellow
Write-Host "5. Change the default PASSWORD immediately after first login" -ForegroundColor Yellow
Write-Host ""

# Verify encryption key length
if ($encryptionKey.Length -eq 32) {  # 32 characters = 32 bytes when converted to []byte in Go
    Write-Host "✓ Encryption key length verified (32 characters = 32 bytes for AES-256)" -ForegroundColor Green
} else {
    Write-Host "✗ Warning: Encryption key length is incorrect! Got $($encryptionKey.Length) chars, expected 32" -ForegroundColor Red
}

Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "1. Review and customize the .env file if needed" -ForegroundColor White
Write-Host "2. Ensure SteamCMD and NSSM are installed and paths are correct" -ForegroundColor White
Write-Host "3. Build and run the application: go run cmd/api/main.go" -ForegroundColor White
Write-Host "4. Change the default admin password on first login" -ForegroundColor White
Write-Host ""
Write-Host "Happy racing! 🏁" -ForegroundColor Green
