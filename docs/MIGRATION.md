# Migration Guide

This guide covers database migrations and version upgrades for ACC Server Manager.

## Database Migrations

The application handles database migrations automatically on startup. No manual intervention is required for most upgrades.

### Automatic Migrations

When you start the application:
1. It checks the current database schema
2. Applies any pending migrations
3. Updates the schema version

### Manual Migration (if needed)

If automatic migration fails:

```bash
# Backup your database first
copy acc.db acc_backup.db

# Delete the database and let it recreate
del acc.db

# Start the application - it will create a fresh database
./api.exe
```

## Upgrading ACC Server Manager

### From v1.x to v2.x

1. **Backup your data**
   ```bash
   copy acc.db acc_backup.db
   copy .env .env.backup
   ```

2. **Stop the application**
   ```bash
   # If running as service
   nssm stop "ACC Server Manager"
   ```

3. **Update the code**
   ```bash
   git pull
   go build -o api.exe cmd/api/main.go
   ```

4. **Update configuration**
   - Check `.env.example` for new required variables
   - Run `.\scripts\generate-secrets.ps1` if needed

5. **Start the application**
   ```bash
   ./api.exe
   # Or restart service
   nssm start "ACC Server Manager"
   ```

## Breaking Changes

### v2.0
- Changed from system_configs table to environment variables
- Now use `STEAMCMD_PATH` and `NSSM_PATH` environment variables
- UUID fields added to all tables (automatic migration)

### v1.5
- Authentication system overhaul
- New permission system
- Password requirements enforced

## Data Backup

### Regular Backups

Create a scheduled task to backup your database:

```powershell
# PowerShell backup script
$date = Get-Date -Format "yyyy-MM-dd"
Copy-Item "acc.db" "backups\acc_$date.db"
```

### What to Backup

- `acc.db` - Main database
- `.env` - Configuration and secrets
- `logs/` - Application logs (optional)
- Server configuration files in each server directory

## Rollback Procedure

If an upgrade fails:

1. Stop the application
2. Restore the database: `copy acc_backup.db acc.db`
3. Restore the configuration: `copy .env.backup .env`
4. Use the previous binary version
5. Start the application

## Common Migration Issues

### "Database locked"
- Stop all instances of the application
- Check for stuck processes

### "Schema version mismatch"
- Let automatic migration complete
- Don't interrupt during migration

### "Missing columns"
- Database migration was interrupted
- Restore from backup and retry

## Best Practices

1. **Always backup before upgrading**
2. **Test upgrades in a non-production environment first**
3. **Read release notes for breaking changes**
4. **Keep the last working version's binary**
5. **Monitor logs during first startup after upgrade**