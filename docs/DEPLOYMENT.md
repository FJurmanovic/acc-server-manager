# Deployment Guide

This guide covers deploying ACC Server Manager to a production Windows server.

## Production Requirements

- Windows Server 2016+ or Windows 10/11
- 4GB+ RAM
- 20GB+ free disk space
- Administrator access
- SteamCMD and NSSM installed

## Deployment Steps

### 1. Prepare the Server

Install required tools:
```powershell
# Create directories
New-Item -ItemType Directory -Path "C:\ACCServerManager"
New-Item -ItemType Directory -Path "C:\steamcmd"
New-Item -ItemType Directory -Path "C:\tools\nssm"

# Download and extract SteamCMD
# Download and extract NSSM
```

### 2. Build for Production

On your development machine:
```bash
# Build optimized binary
go build -ldflags="-w -s" -o acc-server-manager.exe cmd/api/main.go
```

### 3. Deploy Files

Copy to server:
- `acc-server-manager.exe`
- `.env` file (with production secrets)
- `nssm.exe` (if not in PATH)

### 4. Configure Production Environment

Generate production secrets:
```powershell
# On the server
.\scripts\generate-secrets.ps1
```

Edit `.env` for production:
```env
PORT=80
CORS_ALLOWED_ORIGIN=https://yourdomain.com
```

### 5. Install as Windows Service

```powershell
# Using NSSM
nssm install "ACC Server Manager" "C:\ACCServerManager\acc-server-manager.exe"
nssm set "ACC Server Manager" DisplayName "ACC Server Manager"
nssm set "ACC Server Manager" Description "Web management for ACC servers"
nssm set "ACC Server Manager" Start SERVICE_AUTO_START
nssm set "ACC Server Manager" AppDirectory "C:\ACCServerManager"

# Start the service
nssm start "ACC Server Manager"
```

### 6. Configure Firewall

```powershell
# Allow HTTP traffic
New-NetFirewallRule -DisplayName "ACC Server Manager" -Direction Inbound -Protocol TCP -LocalPort 80 -Action Allow
```

### 7. Set Up Reverse Proxy (Optional)

If using IIS as reverse proxy:
1. Install URL Rewrite and ARR modules
2. Configure reverse proxy to localhost:3000
3. Enable HTTPS with valid certificate

## Security Checklist

- [ ] Generated unique production secrets
- [ ] Changed default admin password
- [ ] Configured HTTPS (via reverse proxy)
- [ ] Restricted database file permissions
- [ ] Enabled Windows Firewall
- [ ] Disabled unnecessary ports
- [ ] Set up backup schedule

## Monitoring

### Service Health
```powershell
# Check service status
Get-Service "ACC Server Manager"

# View recent logs
Get-EventLog -LogName Application -Source "ACC Server Manager" -Newest 20
```

### Application Logs
- Check `logs/app.log` for application events
- Check `logs/error.log` for errors
- Monitor disk space for log growth

## Backup Strategy

### Automated Backups

Create scheduled task for daily backups:
```powershell
# Backup script (save as backup.ps1)
$date = Get-Date -Format "yyyy-MM-dd"
$backupDir = "C:\Backups\ACCServerManager"
New-Item -ItemType Directory -Force -Path $backupDir

# Backup database and config
Copy-Item "C:\ACCServerManager\acc.db" "$backupDir\acc_$date.db"
Copy-Item "C:\ACCServerManager\.env" "$backupDir\env_$date"

# Keep only last 7 days
Get-ChildItem $backupDir -File | Where-Object {$_.LastWriteTime -lt (Get-Date).AddDays(-7)} | Remove-Item
```

## Updates

### Update Process

1. **Backup current deployment**
2. **Build new version**
3. **Stop service**: `nssm stop "ACC Server Manager"`
4. **Replace binary**
5. **Start service**: `nssm start "ACC Server Manager"`
6. **Verify**: Check logs and web interface

### Rollback

If update fails:
1. Stop service
2. Restore previous binary
3. Restore database if needed
4. Start service

## Troubleshooting Deployment

### Service Won't Start
- Check Event Viewer for errors
- Verify .env file exists and is valid
- Run manually to see console output

### Can't Access Web Interface
- Check firewall rules
- Verify service is running
- Check port binding in .env

### Permission Errors
- Run service as Administrator (not recommended)
- Or grant specific permissions to service account

## Performance Tuning

### For Large Deployments
- Use SSD for database storage
- Increase Windows TCP connection limits
- Consider load balancing for 50+ servers
- Monitor memory usage and adjust if needed

### Database Maintenance
```sql
-- Run monthly via SQLite
VACUUM;
ANALYZE;
```
