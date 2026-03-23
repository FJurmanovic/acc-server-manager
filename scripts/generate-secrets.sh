#!/bin/bash

# ACC Server Manager - Secret Generation Script (Unix/Linux)
# This script generates cryptographically secure secrets for the ACC Server Manager

echo -e "\033[32mACC Server Manager - Secret Generation Script\033[0m"
echo -e "\033[32m=============================================\033[0m"
echo ""

# Check if openssl is available
if ! command -v openssl &> /dev/null; then
    echo -e "\033[31mError: openssl is required but not installed.\033[0m"
    echo "Please install openssl and try again."
    echo ""
    echo "Ubuntu/Debian: sudo apt-get install openssl"
    echo "CentOS/RHEL:   sudo yum install openssl"
    echo "macOS:         brew install openssl"
    exit 1
fi

# Generate secrets using openssl
echo -e "\033[33mGenerating cryptographically secure secrets...\033[0m"
echo ""

JWT_SECRET=$(openssl rand -base64 64)
APP_SECRET=$(openssl rand -hex 32)
APP_SECRET_CODE=$(openssl rand -hex 32)
ENCRYPTION_KEY=$(openssl rand -hex 16)

# Display generated secrets
echo -e "\033[36mGenerated Secrets:\033[0m"
echo -e "\033[36m==================\033[0m"
echo ""
echo -e "\033[37mJWT_SECRET=\033[0m\033[33m$JWT_SECRET\033[0m"
echo ""
echo -e "\033[37mAPP_SECRET=\033[0m\033[33m$APP_SECRET\033[0m"
echo ""
echo -e "\033[37mAPP_SECRET_CODE=\033[0m\033[33m$APP_SECRET_CODE\033[0m"
echo ""
echo -e "\033[37mENCRYPTION_KEY=\033[0m\033[33m$ENCRYPTION_KEY\033[0m"
echo ""

# Check if .env file exists
ENV_FILE=".env"
if [ -f "$ENV_FILE" ]; then
    echo -e "\033[31mWarning: .env file already exists!\033[0m"
    read -p "Do you want to update it with new secrets? (y/N): " overwrite
    if [[ $overwrite =~ ^[Yy]$ ]]; then
        UPDATE_FILE=true
    else
        UPDATE_FILE=false
        echo -e "\033[33mSecrets generated but not written to file.\033[0m"
    fi
else
    read -p "Create .env file with these secrets? (Y/n): " create_file
    if [[ $create_file =~ ^[Nn]$ ]]; then
        UPDATE_FILE=false
        echo -e "\033[33mSecrets generated but not written to file.\033[0m"
    else
        UPDATE_FILE=true
    fi
fi

if [ "$UPDATE_FILE" = true ]; then
    if [ -f "$ENV_FILE" ]; then
        # Backup existing file
        BACKUP_FILE=".env.backup.$(date +%Y%m%d-%H%M%S)"
        cp "$ENV_FILE" "$BACKUP_FILE"
        echo -e "\033[32mBacked up existing .env to $BACKUP_FILE\033[0m"

        # Update existing file
        if command -v sed &> /dev/null; then
            # Use sed to update secrets in place
            sed -i.tmp "s/^JWT_SECRET=.*/JWT_SECRET=$JWT_SECRET/" "$ENV_FILE"
            sed -i.tmp "s/^APP_SECRET=.*/APP_SECRET=$APP_SECRET/" "$ENV_FILE"
            sed -i.tmp "s/^APP_SECRET_CODE=.*/APP_SECRET_CODE=$APP_SECRET_CODE/" "$ENV_FILE"
            sed -i.tmp "s/^ENCRYPTION_KEY=.*/ENCRYPTION_KEY=$ENCRYPTION_KEY/" "$ENV_FILE"
            rm -f "$ENV_FILE.tmp"
            echo -e "\033[32mUpdated .env file with new secrets\033[0m"
        else
            echo -e "\033[31mError: sed command not found. Please update the .env file manually.\033[0m"
        fi
    else
        # Create new .env file
        if [ -f ".env.example" ]; then
            # Create from template
            cp ".env.example" "$ENV_FILE"
            if command -v sed &> /dev/null; then
                sed -i.tmp "s/^JWT_SECRET=.*/JWT_SECRET=$JWT_SECRET/" "$ENV_FILE"
                sed -i.tmp "s/^APP_SECRET=.*/APP_SECRET=$APP_SECRET/" "$ENV_FILE"
                sed -i.tmp "s/^APP_SECRET_CODE=.*/APP_SECRET_CODE=$APP_SECRET_CODE/" "$ENV_FILE"
                sed -i.tmp "s/^ENCRYPTION_KEY=.*/ENCRYPTION_KEY=$ENCRYPTION_KEY/" "$ENV_FILE"
                rm -f "$ENV_FILE.tmp"
                echo -e "\033[32mCreated .env file from template with generated secrets\033[0m"
            else
                echo -e "\033[31mError: sed command not found. Please update the .env file manually.\033[0m"
            fi
        else
            # Create minimal .env file
            cat > "$ENV_FILE" << EOF
# ACC Server Manager Environment Configuration
# Generated on $(date '+%Y-%m-%d %H:%M:%S')

# CRITICAL SECURITY SETTINGS (REQUIRED)
JWT_SECRET=$JWT_SECRET
APP_SECRET=$APP_SECRET
APP_SECRET_CODE=$APP_SECRET_CODE
ENCRYPTION_KEY=$ENCRYPTION_KEY

# CORE APPLICATION SETTINGS
DB_NAME=acc.db
PORT=3000
CORS_ALLOWED_ORIGIN=http://localhost:5173
PASSWORD=change-this-default-admin-password
EOF
            echo -e "\033[32mCreated minimal .env file with generated secrets\033[0m"
        fi
    fi
fi

echo ""
echo -e "\033[31mSecurity Notes:\033[0m"
echo -e "\033[31m===============\033[0m"
echo -e "\033[33m1. Keep these secrets secure and never commit them to version control\033[0m"
echo -e "\033[33m2. Use different secrets for each environment (dev, staging, production)\033[0m"
echo -e "\033[33m3. Rotate secrets regularly in production environments\033[0m"
echo -e "\033[33m4. The ENCRYPTION_KEY is exactly 32 bytes as required for AES-256\033[0m"
echo -e "\033[33m5. Change the default PASSWORD immediately after first login\033[0m"
echo ""

# Verify encryption key length (32 characters = 32 bytes when converted to []byte in Go)
if [ ${#ENCRYPTION_KEY} -eq 32 ]; then
    echo -e "\033[32m✓ Encryption key length verified (32 characters = 32 bytes for AES-256)\033[0m"
else
    echo -e "\033[31m✗ Warning: Encryption key length is incorrect!\033[0m"
fi

echo ""
echo -e "\033[36mNext steps:\033[0m"
echo -e "\033[37m1. Review and customize the .env file if needed\033[0m"
echo -e "\033[37m2. Install Go 1.23.0 or later if not already installed\033[0m"
echo -e "\033[37m3. Build and run the application: go run cmd/api/main.go\033[0m"
echo -e "\033[37m4. Change the default admin password on first login\033[0m"
echo ""
echo -e "\033[32mHappy racing! 🏁\033[0m"
