# Environment Variables Configuration

This document describes the environment variables used by the ACC Server Manager to replace the previous database-based system configuration.

## Overview

The `system_configs` database table has been completely removed and replaced with environment variables for better configuration management and deployment flexibility.

## Environment Variables

### STEAMCMD_PATH
**Description:** Path to the SteamCMD executable  
**Default:** `c:\steamcmd\steamcmd.exe`  
**Example:** `STEAMCMD_PATH=D:\tools\steamcmd\steamcmd.exe`

This path is used for:
- Installing ACC dedicated servers
- Updating server files
- Managing Steam-based server installations

### NSSM_PATH
**Description:** Path to the NSSM (Non-Sucking Service Manager) executable  
**Default:** `.\nssm.exe`  
**Example:** `NSSM_PATH=C:\tools\nssm\win64\nssm.exe`

This path is used for:
- Creating Windows services for ACC servers
- Managing service lifecycle (start, stop, restart)
- Service configuration and management

## Setting Environment Variables

### Windows Command Prompt
```cmd
set STEAMCMD_PATH=D:\tools\steamcmd\steamcmd.exe
set NSSM_PATH=C:\tools\nssm\win64\nssm.exe
```

### Windows PowerShell
```powershell
$env:STEAMCMD_PATH = "D:\tools\steamcmd\steamcmd.exe"
$env:NSSM_PATH = "C:\tools\nssm\win64\nssm.exe"
```

### System Environment Variables (Persistent)
1. Open System Properties → Advanced → Environment Variables
2. Add new system variables:
   - Variable name: `STEAMCMD_PATH`
   - Variable value: `D:\tools\steamcmd\steamcmd.exe`
3. Repeat for `NSSM_PATH`

### Docker Environment
```dockerfile
ENV STEAMCMD_PATH=/opt/steamcmd/steamcmd.sh
ENV NSSM_PATH=/usr/local/bin/nssm
```

### Docker Compose
```yaml
environment:
  - STEAMCMD_PATH=/opt/steamcmd/steamcmd.sh
  - NSSM_PATH=/usr/local/bin/nssm
```

## Migration from system_configs

### Automatic Migration
A migration script (`003_remove_system_configs.sql`) will automatically:
1. Remove the `system_configs` table
2. Clean up related database references
3. Record the migration in `migration_records`

### Manual Configuration Required
After upgrading, you must set the environment variables based on your previous system configuration:

1. Check your previous configuration (if you had custom paths):
   ```sql
   SELECT key, value, default_value FROM system_configs;
   ```

2. Set environment variables accordingly:
   - If you used custom `steamcmd_path`: Set `STEAMCMD_PATH`
   - If you used custom `nssm_path`: Set `NSSM_PATH`

3. Restart the ACC Server Manager service

### Validation
The application will use default values if environment variables are not set. To validate your configuration:

1. Check the application logs on startup
2. The `env.ValidatePaths()` function can be used to verify paths exist
3. Monitor for any "failed to get path" errors in logs

## Benefits of Environment Variables

### Deployment Flexibility
- Different environments can have different tool paths
- No database dependency for basic configuration
- Container-friendly configuration

### Security
- Sensitive paths not stored in database
- Environment-specific configuration
- Better separation of configuration from data

### Performance
- No database queries for basic path lookups
- Reduced database load on every operation
- Faster service startup

## Troubleshooting

### Common Issues

**Issue:** SteamCMD operations fail  
**Solution:** Verify `STEAMCMD_PATH` points to valid steamcmd.exe

**Issue:** Service creation fails  
**Solution:** Verify `NSSM_PATH` points to valid nssm.exe

**Issue:** Using default paths  
**Solution:** Set environment variables and restart application

### Debugging
Enable debug logging to see which paths are being used:
```
2024-01-01 12:00:00 DEBUG Using SteamCMD path: D:\tools\steamcmd\steamcmd.exe
2024-01-01 12:00:00 DEBUG Using NSSM path: C:\tools\nssm\win64\nssm.exe
```

## Code Changes Summary

### Removed Components
- `local/model/config.go` - SystemConfig struct and related constants
- `local/service/system_config_service.go` - SystemConfigService
- `local/repository/system_config_repository.go` - SystemConfigRepository
- Database table: `system_configs`

### Added Components
- `local/utl/env/env.go` - Environment variable utilities
- Migration script: `003_remove_system_configs.sql`

### Modified Services
- **SteamService**: Now uses `env.GetSteamCMDPath()`
- **WindowsService**: Now uses `env.GetNSSMPath()`
- **ServerService**: Removed SystemConfigService dependency
- **ApiService**: Removed SystemConfigService dependency