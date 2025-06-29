# Configuration Guide for ACC Server Manager

## Overview

This guide provides comprehensive information about configuring the ACC Server Manager application, including environment variables, server settings, security configurations, and advanced options.

## 📁 Configuration Files

### Environment Configuration (.env)

The primary configuration is handled through environment variables. Create a `.env` file in the root directory:

```bash
# Copy the example file
cp .env.example .env
```

### Configuration File Hierarchy

1. **Environment Variables** (highest priority)
2. **`.env` file** (medium priority)
3. **Default values** (lowest priority)

## 🔐 Security Configuration

### Required Security Variables

These variables are **mandatory** and the application will not start without them:

```env
# JWT Secret - Used for signing authentication tokens
# Generate with: openssl rand -base64 64
JWT_SECRET=your-super-secure-jwt-secret-minimum-64-characters-long

# Application Secrets - Used for internal encryption and security
# Generate with: openssl rand -hex 32
APP_SECRET=your-32-character-hex-secret-here
APP_SECRET_CODE=your-32-character-hex-secret-code-here

# Encryption Key - Used for AES-256 encryption (MUST be exactly 32 bytes)
# Generate with: openssl rand -hex 32
ENCRYPTION_KEY=your-exactly-32-byte-encryption-key-here
```

## 🌐 Server Configuration

### Basic Server Settings

```env
# HTTP server port
PORT=3000

# CORS allowed origin (comma-separated for multiple origins)
CORS_ALLOWED_ORIGIN=http://localhost:5173

# Database file name (SQLite)
DB_NAME=acc.db

# Default admin password for initial setup (change after first login)
PASSWORD=change-this-default-admin-password
```

**Note**: Most other server configuration options (timeouts, request limits, etc.) are handled by application defaults and don't require environment variables.

## 🗄️ Database Configuration

The application uses SQLite with minimal configuration required:

```env
# Database file name (only setting available via environment)
DB_NAME=acc.db
```

**Note**: Other database settings like connection timeouts, migration settings, and SQL logging are handled internally by the application and don't require environment variables.

## 🎮 Steam Integration

Steam integration settings are managed through the web interface and stored in the database as system configuration. No environment variables are required for Steam integration.

**Configuration via Web Interface:**
- SteamCMD executable path
- NSSM executable path  
- Steam credentials (encrypted in database)
- Update schedules and preferences

**Default Values:**
- SteamCMD Path: `c:\steamcmd\steamcmd.exe`
- NSSM Path: `.\nssm.exe`

## 🔥 Windows Service Configuration

Windows service and firewall configurations are handled internally by the application:

**Service Management:**
- NSSM path configured via web interface
- Default service name prefix: `ACC-Server`
- Automatic service creation and management

**Firewall Management:**
- Automatic firewall rule creation
- Default TCP port range: 9600+
- Default UDP port range: 9600+
- Rule cleanup on server deletion

**No environment variables required** - all settings are managed through the system configuration interface.

## 📊 Logging Configuration

Logging is handled internally by the application with sensible defaults:

**Default Logging Behavior:**
- Log level: `info` (adjustable via code)
- Log format: Structured text format
- Log files: Automatic rotation and cleanup
- Security events: Automatically logged
- Error tracking: Comprehensive error logging

**No environment variables required** - logging configuration is built into the application.

## 🚦 Rate Limiting Configuration

Rate limiting is built into the application with secure defaults:

**Built-in Rate Limits:**
- Global: 100 requests per minute per IP
- Authentication: 5 attempts per 15 minutes per IP
- API endpoints: 60 requests per minute per IP
- Configuration updates: Protected with additional limits

**No environment variables required** - rate limiting is automatically applied with appropriate limits for security and performance.

## 📈 Monitoring Configuration

Monitoring features are built into the application:

**Available Monitoring:**
- Health check endpoint: `/health` (always enabled)
- Performance tracking: Built-in performance monitoring
- Error tracking: Automatic error logging and tracking
- Security monitoring: Authentication and authorization events

**No environment variables required** - monitoring is automatically enabled with appropriate defaults.

## 🔄 Backup Configuration

Backup functionality is handled internally:

**Automatic Backups:**
- Database backup before migrations
- Configuration file versioning
- Error recovery mechanisms

**Manual Backups:**
- Database files can be copied manually
- Configuration export/import via web interface

**No environment variables required** - backup features are built into the application workflow.

## 🧪 Development Configuration

### Development Mode Settings

```env
# Enable development mode (NEVER use in production)
DEV_MODE=false

# Enable debug endpoints
DEBUG_ENDPOINTS=false

# Enable hot reload (requires air)
HOT_RELOAD=false

# Disable security features for testing (DANGEROUS)
DISABLE_SECURITY=false
```

### Testing Configuration

```env
# Test database name
TEST_DB_NAME=acc_test.db

# Enable test fixtures
ENABLE_TEST_FIXTURES=false

# Test timeout in seconds
TEST_TIMEOUT=300
```

## 🏭 Production Configuration

### Production Deployment Settings

```env
# Production mode
PRODUCTION=true

# Enable HTTPS enforcement
FORCE_HTTPS=true

# Security-first configuration
SECURITY_STRICT=true

# Disable debug information
DISABLE_DEBUG_INFO=true

# Enable comprehensive monitoring
ENABLE_MONITORING=true
```

### Performance Optimization

```env
# Enable response compression
ENABLE_COMPRESSION=true

# Compression level (1-9)
COMPRESSION_LEVEL=6

# Enable response caching
ENABLE_CACHING=true

# Cache TTL in seconds
CACHE_TTL=300

# Maximum cache size in MB
CACHE_MAX_SIZE=100
```

## 🛠️ Advanced Configuration

### Custom Port Ranges

```env
# Custom TCP port ranges (comma-separated)
CUSTOM_TCP_PORTS=9600-9610,9700-9710

# Custom UDP port ranges (comma-separated)
CUSTOM_UDP_PORTS=9600-9610,9700-9710

# Exclude specific ports (comma-separated)
EXCLUDED_PORTS=9605,9705
```

### Custom Paths

```env
# Custom ACC server installation path
ACC_SERVER_PATH=C:\ACC_Server

# Custom configuration templates path
CONFIG_TEMPLATES_PATH=./templates

# Custom scripts path
SCRIPTS_PATH=./scripts
```

### Integration Settings

```env
# External API endpoints
EXTERNAL_API_ENABLED=false
EXTERNAL_API_URL=https://api.example.com
EXTERNAL_API_KEY=your-api-key-here

# Webhook notifications
WEBHOOK_ENABLED=false
WEBHOOK_URL=https://your-webhook-url.com
WEBHOOK_SECRET=your-webhook-secret
```

## 📋 Configuration Validation

### Validation Rules

The application automatically validates configuration on startup:

1. **Required Variables**: Must be present and non-empty
2. **Numeric Values**: Must be valid numbers within acceptable ranges
3. **File Paths**: Must be accessible and have appropriate permissions
4. **URLs**: Must be valid URL format
5. **Encryption Keys**: Must be exactly 32 bytes for AES-256

### Configuration Errors

Common configuration errors and solutions:

#### "JWT_SECRET must be at least 32 bytes long"
- **Solution**: Generate a longer JWT secret using `openssl rand -base64 64`

#### "ENCRYPTION_KEY must be exactly 32 bytes long"
- **Solution**: Generate a 32-byte key using `openssl rand -hex 32`

#### "Invalid port number"
- **Solution**: Ensure port numbers are between 1 and 65535

#### "SteamCMD not found"
- **Solution**: Install SteamCMD and update the `STEAMCMD_PATH` variable

## 🔧 Configuration Management

### Environment-Specific Configurations

#### Development (.env.development)
```env
DEV_MODE=true
LOG_LEVEL=debug
CORS_ALLOWED_ORIGIN=http://localhost:3000
```

#### Production (.env.production)
```env
PRODUCTION=true
FORCE_HTTPS=true
LOG_LEVEL=warn
SECURITY_STRICT=true
```

#### Testing (.env.test)
```env
DB_NAME=acc_test.db
LOG_LEVEL=error
DISABLE_RATE_LIMITING=true
```

### Configuration Templates

Create configuration templates for common setups:

#### Single Server Setup
```env
# Minimal configuration for single server
PORT=3000
DB_NAME=acc.db
SERVICE_NAME_PREFIX=ACC-Server
```

#### Multi-Server Setup
```env
# Configuration for multiple servers
AUTO_FIREWALL_RULES=true
PORT_RANGE_SIZE=20
SERVICE_START_TIMEOUT=120
```

#### High-Security Setup
```env
# Maximum security configuration
FORCE_HTTPS=true
RATE_LIMIT_AUTH=3
SESSION_TIMEOUT=30
MAX_LOGIN_ATTEMPTS=3
LOCKOUT_DURATION=30
SECURITY_STRICT=true
```

## 🚨 Security Best Practices

### Secret Management

1. **Never commit secrets to version control**
2. **Use environment-specific secret files**
3. **Rotate secrets regularly**
4. **Use secure secret generation methods**
5. **Limit access to configuration files**

### Production Security

1. **Enable HTTPS enforcement**
2. **Configure appropriate CORS origins**
3. **Set up proper rate limiting**
4. **Enable comprehensive logging**
5. **Regular security audits**

## 📞 Configuration Support

### Troubleshooting

For configuration issues:

1. Check the application logs for specific error messages
2. Validate environment variables using the built-in validation
3. Refer to the examples in `.env.example`
4. Test configuration changes in development first

### Getting Help

- **Documentation**: Check this guide and other documentation files
- **Issues**: Report configuration bugs via GitHub Issues
- **Community**: Ask questions in community discussions
- **Professional Support**: Contact maintainers for enterprise support

---

**Note**: Always test configuration changes in a development environment before applying them to production.