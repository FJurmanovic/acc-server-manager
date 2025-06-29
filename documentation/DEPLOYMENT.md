# Deployment Guide for ACC Server Manager

## Overview

This guide provides comprehensive instructions for deploying the ACC Server Manager in various environments, from development to production. It covers security considerations, performance optimization, monitoring setup, and maintenance procedures.

## 🚀 Quick Start Deployment

### Prerequisites Checklist

- [ ] Windows 10/11 or Windows Server 2016+
- [ ] Go 1.23.0 or later installed
- [ ] Administrative privileges
- [ ] Valid Steam account
- [ ] Internet connection for Steam downloads

### Minimum System Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| **CPU** | 2 cores | 4+ cores |
| **RAM** | 4 GB | 8+ GB |
| **Storage** | 10 GB free | 50+ GB SSD |
| **Network** | 10 Mbps | 100+ Mbps |

## 📦 Installation Methods

### Method 1: Binary Deployment (Recommended)

1. **Download Release Binary**
   ```bash
   # Download the latest release from GitHub
   # Extract to your installation directory
   cd C:\ACC-Server-Manager
   ```

2. **Configure Environment**
   ```bash
   copy .env.example .env
   # Edit .env with your configuration
   ```

3. **Generate Secrets**
   ```bash
   # Generate JWT secret
   openssl rand -base64 64
   
   # Generate app secrets
   openssl rand -hex 32
   
   # Generate encryption key
   openssl rand -hex 32
   ```

4. **Run Application**
   ```bash
   .\acc-server-manager.exe
   ```

### Method 2: Source Code Deployment

1. **Clone Repository**
   ```bash
   git clone https://github.com/FJurmanovic/acc-server-manager.git
   cd acc-server-manager
   ```

2. **Install Dependencies**
   ```bash
   go mod download
   go mod verify
   ```

3. **Build Application**
   ```bash
   # Development build
   go build -o acc-server-manager.exe cmd/api/main.go
   
   # Production build (optimized)
   go build -ldflags="-w -s" -o acc-server-manager.exe cmd/api/main.go
   ```

4. **Configure and Run**
   ```bash
   copy .env.example .env
   # Configure your .env file
   .\acc-server-manager.exe
   ```

## 🔧 Environment Configuration

### Production Environment Variables

Create a production `.env` file:

```env
# ========================================
# PRODUCTION CONFIGURATION
# ========================================

# Security (REQUIRED - Generate unique values)
JWT_SECRET=your-production-jwt-secret-64-chars-minimum
APP_SECRET=your-production-app-secret-32-chars
APP_SECRET_CODE=your-production-secret-code-32-chars
ENCRYPTION_KEY=your-production-encryption-key-32-bytes

# Server Configuration
PORT=8080
HOST=0.0.0.0
PRODUCTION=true
FORCE_HTTPS=true

# Database
DB_NAME=acc_production.db
DB_PATH=./data

# CORS (Set to your actual domain)
CORS_ALLOWED_ORIGIN=https://yourdomain.com

# Security Settings
RATE_LIMIT_GLOBAL=1000
RATE_LIMIT_AUTH=10
SESSION_TIMEOUT=120
MAX_LOGIN_ATTEMPTS=5
LOCKOUT_DURATION=30

# Steam Configuration
STEAMCMD_PATH=C:\steamcmd\steamcmd.exe
NSSM_PATH=C:\nssm\nssm.exe

# Logging
LOG_LEVEL=warn
LOG_FILE=./logs/production.log
LOG_MAX_SIZE=100
LOG_MAX_FILES=10

# Monitoring
HEALTH_CHECK_ENABLED=true
METRICS_ENABLED=true
PERFORMANCE_MONITORING=true

# Backup
AUTO_BACKUP=true
BACKUP_INTERVAL=12
BACKUP_RETENTION=30
BACKUP_DIR=./backups
```

### Development Environment Variables

```env
# ========================================
# DEVELOPMENT CONFIGURATION
# ========================================

# Security (Use secure values even in dev)
JWT_SECRET=dev-jwt-secret-but-still-secure-64-chars-minimum
APP_SECRET=dev-app-secret-32-chars-here
APP_SECRET_CODE=dev-secret-code-32-chars-here
ENCRYPTION_KEY=dev-encryption-key-32-bytes-here

# Server Configuration
PORT=3000
HOST=localhost
DEV_MODE=true
DEBUG_ENDPOINTS=true

# Database
DB_NAME=acc_dev.db

# CORS
CORS_ALLOWED_ORIGIN=http://localhost:3000,http://localhost:5173

# Relaxed Security (Development Only)
RATE_LIMIT_GLOBAL=1000
DISABLE_SECURITY=false

# Logging
LOG_LEVEL=debug
LOG_COLORS=true
ENABLE_SQL_LOGGING=true

# Development Tools
HOT_RELOAD=true
ENABLE_TEST_FIXTURES=true
```

## 🔒 Security Hardening

### SSL/TLS Configuration

1. **Obtain SSL Certificate**
   ```bash
   # Option 1: Let's Encrypt (Free)
   certbot certonly --webroot -w /var/www/html -d yourdomain.com
   
   # Option 2: Commercial Certificate
   # Purchase and install certificate from CA
   ```

2. **Configure Reverse Proxy (Nginx)**
   ```nginx
   server {
       listen 443 ssl http2;
       server_name yourdomain.com;
       
       ssl_certificate /path/to/certificate.crt;
       ssl_certificate_key /path/to/private.key;
       ssl_protocols TLSv1.2 TLSv1.3;
       ssl_ciphers ECDHE+AESGCM:ECDHE+AES256:ECDHE+AES128:!aNULL:!MD5:!DSS;
       
       location / {
           proxy_pass http://localhost:8080;
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
           proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
           proxy_set_header X-Forwarded-Proto $scheme;
       }
   }
   
   # Redirect HTTP to HTTPS
   server {
       listen 80;
       server_name yourdomain.com;
       return 301 https://$server_name$request_uri;
   }
   ```

3. **Configure Application for SSL**
   ```env
   FORCE_HTTPS=true
   CORS_ALLOWED_ORIGIN=https://yourdomain.com
   ```

### Firewall Configuration

1. **Windows Firewall Rules**
   ```powershell
   # Allow application through Windows Firewall
   New-NetFirewallRule -DisplayName "ACC Server Manager" -Direction Inbound -Protocol TCP -LocalPort 8080 -Action Allow
   
   # Allow ACC server ports (adjust range as needed)
   New-NetFirewallRule -DisplayName "ACC Servers TCP" -Direction Inbound -Protocol TCP -LocalPort 9600-9700 -Action Allow
   New-NetFirewallRule -DisplayName "ACC Servers UDP" -Direction Inbound -Protocol UDP -LocalPort 9600-9700 -Action Allow
   ```

2. **Network Security Groups (Azure)**
   ```json
   {
     "securityRules": [
       {
         "name": "AllowHTTPS",
         "properties": {
           "protocol": "TCP",
           "sourcePortRange": "*",
           "destinationPortRange": "443",
           "sourceAddressPrefix": "*",
           "destinationAddressPrefix": "*",
           "access": "Allow",
           "priority": 1000,
           "direction": "Inbound"
         }
       }
     ]
   }
   ```

### User Access Control

1. **Create Dedicated Service Account**
   ```powershell
   # Create service account
   New-LocalUser -Name "ACCServiceUser" -Description "ACC Server Manager Service Account" -NoPassword
   Add-LocalGroupMember -Group "Users" -Member "ACCServiceUser"
   
   # Set permissions on application directory
   icacls "C:\ACC-Server-Manager" /grant "ACCServiceUser:(OI)(CI)F"
   ```

2. **Configure Service Permissions**
   ```powershell
   # Grant service logon rights
   secedit /export /cfg security.inf
   # Edit security.inf to add ACCServiceUser to SeServiceLogonRight
   secedit /configure /db security.sdb /cfg security.inf
   ```

## 🏗️ Service Installation

### Windows Service with NSSM

1. **Install NSSM**
   ```bash
   # Download NSSM from https://nssm.cc/
   # Extract nssm.exe to C:\nssm\
   ```

2. **Create Service**
   ```powershell
   # Install service
   C:\nssm\nssm.exe install "ACCServerManager" "C:\ACC-Server-Manager\acc-server-manager.exe"
   
   # Configure service
   C:\nssm\nssm.exe set "ACCServerManager" DisplayName "ACC Server Manager"
   C:\nssm\nssm.exe set "ACCServerManager" Description "Assetto Corsa Competizione Server Manager"
   C:\nssm\nssm.exe set "ACCServerManager" Start SERVICE_AUTO_START
   C:\nssm\nssm.exe set "ACCServerManager" AppDirectory "C:\ACC-Server-Manager"
   C:\nssm\nssm.exe set "ACCServerManager" ObjectName ".\ACCServiceUser" "password"
   
   # Configure logging
   C:\nssm\nssm.exe set "ACCServerManager" AppStdout "C:\ACC-Server-Manager\logs\service.log"
   C:\nssm\nssm.exe set "ACCServerManager" AppStderr "C:\ACC-Server-Manager\logs\service-error.log"
   
   # Start service
   C:\nssm\nssm.exe start "ACCServerManager"
   ```

3. **Service Management**
   ```powershell
   # Check service status
   Get-Service -Name "ACCServerManager"
   
   # Start/Stop service
   Start-Service -Name "ACCServerManager"
   Stop-Service -Name "ACCServerManager"
   
   # Remove service (if needed)
   C:\nssm\nssm.exe remove "ACCServerManager" confirm
   ```

### Systemd Service (Linux/WSL)

```ini
[Unit]
Description=ACC Server Manager
After=network.target

[Service]
Type=simple
User=accmanager
WorkingDirectory=/opt/acc-server-manager
ExecStart=/opt/acc-server-manager/acc-server-manager
Restart=always
RestartSec=10
Environment=PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
EnvironmentFile=/opt/acc-server-manager/.env

[Install]
WantedBy=multi-user.target
```

## 📊 Monitoring Setup

### Health Check Monitoring

1. **Configure Health Checks**
   ```env
   HEALTH_CHECK_ENABLED=true
   HEALTH_CHECK_PATH=/health
   HEALTH_CHECK_TIMEOUT=10
   ```

2. **External Monitoring (UptimeRobot)**
   ```bash
   # Monitor endpoint: https://yourdomain.com/health
   # Expected response: 200 OK with JSON health status
   ```

### Log Management

1. **Log Rotation Configuration**
   ```env
   LOG_MAX_SIZE=100
   LOG_MAX_FILES=10
   LOG_MAX_AGE=30
   ```

2. **Centralized Logging (Optional)**
   ```yaml
   # docker-compose.yml for ELK Stack
   version: '3'
   services:
     elasticsearch:
       image: elasticsearch:7.14.0
     logstash:
       image: logstash:7.14.0
     kibana:
       image: kibana:7.14.0
   ```

### Performance Monitoring

1. **Enable Metrics**
   ```env
   METRICS_ENABLED=true
   METRICS_PORT=9090
   PERFORMANCE_MONITORING=true
   ```

2. **Prometheus Configuration**
   ```yaml
   # prometheus.yml
   global:
     scrape_interval: 15s
   
   scrape_configs:
     - job_name: 'acc-server-manager'
       static_configs:
         - targets: ['localhost:9090']
   ```

## 🔄 Database Management

### Database Backup Strategy

1. **Automated Backups**
   ```env
   AUTO_BACKUP=true
   BACKUP_INTERVAL=12
   BACKUP_RETENTION=30
   BACKUP_DIR=./backups
   BACKUP_COMPRESS=true
   ```

2. **Manual Backup**
   ```powershell
   # Create manual backup
   $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
   Copy-Item "acc.db" "backups/acc-backup-$timestamp.db"
   
   # Compress backup
   Compress-Archive "backups/acc-backup-$timestamp.db" "backups/acc-backup-$timestamp.zip"
   ```

3. **Backup Verification**
   ```bash
   # Test backup integrity
   sqlite3 backup.db "PRAGMA integrity_check;"
   ```

### Database Migration

1. **Pre-Migration Backup**
   ```bash
   # Always backup before migration
   copy acc.db acc-pre-migration-backup.db
   ```

2. **Migration Process**
   ```bash
   # Migration runs automatically on startup
   # Check logs for migration status
   tail -f logs/app.log | grep -i migration
   ```

## 🌐 Load Balancing (High Availability)

### Multiple Instance Setup

1. **Load Balancer Configuration (HAProxy)**
   ```haproxy
   global
       daemon
   
   defaults
       mode http
       timeout connect 5000ms
       timeout client 50000ms
       timeout server 50000ms
   
   frontend acc_frontend
       bind *:80
       default_backend acc_servers
   
   backend acc_servers
       balance roundrobin
       server acc1 10.0.0.10:8080 check
       server acc2 10.0.0.11:8080 check
       server acc3 10.0.0.12:8080 check
   ```

2. **Shared Database Setup**
   ```bash
   # Use network-attached storage for database
   # Mount shared volume on all instances
   net use Z: \\fileserver\acc-shared
   ```

### Session Clustering

```env
# Redis for session storage
REDIS_URL=redis://localhost:6379
SESSION_STORE=redis
```

## 🔧 Maintenance Procedures

### Regular Maintenance Tasks

1. **Daily Tasks**
   ```powershell
   # Check service status
   Get-Service -Name "ACCServerManager"
   
   # Check disk space
   Get-WmiObject -Class Win32_LogicalDisk | Select-Object DeviceID, Size, FreeSpace
   
   # Review error logs
   Get-Content "logs/error.log" -Tail 50
   ```

2. **Weekly Tasks**
   ```powershell
   # Update system patches
   Install-Module PSWindowsUpdate
   Get-WUInstall -AcceptAll -AutoReboot
   
   # Clean old log files
   Get-ChildItem "logs\" -Name "*.log.*" | Where-Object {$_.LastWriteTime -lt (Get-Date).AddDays(-30)} | Remove-Item
   
   # Verify backup integrity
   sqlite3 backups/latest.db "PRAGMA integrity_check;"
   ```

3. **Monthly Tasks**
   ```powershell
   # Update dependencies
   go get -u ./...
   go mod tidy
   
   # Security scan
   go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
   gosec ./...
   
   # Performance review
   # Review metrics and optimize based on usage patterns
   ```

### Update Procedures

1. **Backup Current Installation**
   ```bash
   # Stop service
   Stop-Service -Name "ACCServerManager"
   
   # Backup application
   Copy-Item -Recurse "C:\ACC-Server-Manager" "C:\ACC-Server-Manager-Backup-$(Get-Date -Format 'yyyyMMdd')"
   ```

2. **Deploy New Version**
   ```bash
   # Download new version
   # Replace executable
   # Update configuration if needed
   
   # Start service
   Start-Service -Name "ACCServerManager"
   ```

3. **Rollback Procedure**
   ```bash
   # Stop service
   Stop-Service -Name "ACCServerManager"
   
   # Restore backup
   Remove-Item -Recurse "C:\ACC-Server-Manager"
   Copy-Item -Recurse "C:\ACC-Server-Manager-Backup-$(Get-Date -Format 'yyyyMMdd')" "C:\ACC-Server-Manager"
   
   # Start service
   Start-Service -Name "ACCServerManager"
   ```

## 🐛 Troubleshooting

### Common Issues

1. **Service Won't Start**
   ```powershell
   # Check service status
   Get-Service -Name "ACCServerManager"
   
   # Check service logs
   Get-Content "logs/service-error.log" -Tail 50
   
   # Check Windows Event Log
   Get-EventLog -LogName System -Source "ACCServerManager" -Newest 10
   ```

2. **Database Connection Issues**
   ```bash
   # Check database file permissions
   icacls acc.db
   
   # Test database connection
   sqlite3 acc.db ".tables"
   
   # Check for database locks
   lsof acc.db  # Linux
   ```

3. **Steam Integration Issues**
   ```bash
   # Verify SteamCMD installation
   C:\steamcmd\steamcmd.exe +quit
   
   # Check Steam credentials
   # Review Steam-related logs
   ```

### Performance Issues

1. **High CPU Usage**
   ```bash
   # Check for resource-intensive operations
   # Monitor process performance
   Get-Process -Name "acc-server-manager" | Select-Object CPU, WorkingSet
   ```

2. **Memory Leaks**
   ```bash
   # Monitor memory usage over time
   # Enable detailed memory profiling
   go tool pprof http://localhost:8080/debug/pprof/heap
   ```

3. **Database Performance**
   ```sql
   -- Analyze database performance
   PRAGMA table_info(servers);
   EXPLAIN QUERY PLAN SELECT * FROM servers WHERE status = 'running';
   ```

## 📞 Support and Resources

### Documentation Resources
- [README.md](../README.md) - Getting started guide
- [SECURITY.md](SECURITY.md) - Security guidelines
- [API.md](API.md) - API documentation
- [CONFIGURATION.md](CONFIGURATION.md) - Configuration reference

### Community Support
- **GitHub Issues** - Bug reports and feature requests
- **Discord Community** - Real-time community support
- **Wiki** - Community-maintained documentation

### Professional Support
- **Enterprise Support** - Professional deployment assistance
- **Consulting Services** - Custom implementation and optimization
- **Training** - Team training and best practices

### Emergency Contacts
```
Production Issues: support@yourdomain.com
Security Issues: security@yourdomain.com
Emergency Hotline: +1-XXX-XXX-XXXX
```

## 📋 Deployment Checklist

### Pre-Deployment
- [ ] System requirements verified
- [ ] Dependencies installed
- [ ] Secrets generated and secured
- [ ] Configuration reviewed
- [ ] Security hardening applied
- [ ] Backup strategy implemented
- [ ] Monitoring configured

### Post-Deployment
- [ ] Service running successfully
- [ ] Health checks passing
- [ ] Logs being written correctly
- [ ] Database accessible
- [ ] API endpoints responding
- [ ] Frontend integration working
- [ ] Monitoring alerts configured
- [ ] Documentation updated

### Production Readiness
- [ ] SSL/TLS configured
- [ ] Firewall rules applied
- [ ] Performance monitoring active
- [ ] Backup procedures tested
- [ ] Update procedures documented
- [ ] Disaster recovery plan created
- [ ] Team training completed

---

**Remember**: Always test deployments in a staging environment before applying to production!