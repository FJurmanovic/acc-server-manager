# Configuration Guide

This guide covers the configuration options for ACC Server Manager.

## Environment Variables

### Required Variables (Auto-generated)

These are automatically created by the setup script:

- `JWT_SECRET` - Authentication token secret
- `APP_SECRET` - Application encryption key
- `ENCRYPTION_KEY` - Database encryption key

### Optional Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Web server port | `3000` |
| `DB_NAME` | Database filename | `acc.db` |
| `STEAMCMD_PATH` | Path to SteamCMD | `c:\steamcmd\steamcmd.exe` |
| `NSSM_PATH` | Path to NSSM | `.\nssm.exe` |
| `CORS_ALLOWED_ORIGIN` | Allowed CORS origins | `http://localhost:5173` |

## Setting Environment Variables

### Temporary (Current Session)

```powershell
# PowerShell
$env:PORT = "8080"
$env:STEAMCMD_PATH = "D:\tools\steamcmd\steamcmd.exe"
```

### Permanent (System-wide)

1. Open System Properties → Advanced → Environment Variables
2. Add new system variables with desired values
3. Restart the application

## Server Configuration

### ACC Server Settings

Each ACC server instance has its own configuration managed through the web interface:

- **Server Name** - Display name in the manager
- **Port Settings** - TCP/UDP ports (auto-assigned or manual)
- **Configuration Files** - Edit `configuration.json`, `settings.json`, etc.

### Firewall Rules

The application automatically manages Windows Firewall rules for ACC servers:

- Creates inbound rules for TCP and UDP ports
- Names rules as "ACC Server - [ServerName]"
- Removes rules when server is deleted

## Security Configuration

### Password Requirements

- Minimum 8 characters
- Mix of uppercase, lowercase, numbers
- Special characters recommended

### Session Management

- JWT tokens expire after 24 hours
- Refresh tokens available for extended sessions
- Configurable timeout in future releases

## Database

### SQLite Configuration

- Database file: `acc.db` (configurable via `DB_NAME`)
- Automatic backups: Not yet implemented
- Location: Application root directory

### Data Encryption

Sensitive data is encrypted using AES-256:

- Steam credentials
- User passwords (bcrypt)
- API keys

## Logging

### Log Files

Located in `logs/` directory:

- `app.log` - General application logs
- `error.log` - Error messages
- `access.log` - HTTP access logs

### Log Rotation

Currently manual - delete old logs periodically.

## Performance Tuning

### Database Optimization

- Use SSD for database location
- Regular VACUUM operations recommended
- Keep database size under 1GB

### Memory Usage

- Base usage: ~50MB
- Per server instance: ~10MB
- Caching: ~100MB

## Advanced Configuration

### Custom Ports

To use custom port ranges for ACC servers:

1. Log into web interface
2. Go to Settings → Server Defaults
3. Set port range (e.g., 9600-9700)

### Multiple IPs

If your server has multiple network interfaces:

1. ACC servers will bind to all interfaces by default
2. Configure specific IPs in ACC server settings files

## Configuration Best Practices

1. **Backup your `.env` file** - Contains encryption keys
2. **Use strong passwords** - Especially for admin account
3. **Regular updates** - Keep ACC Server Manager updated
4. **Monitor logs** - Check for errors or warnings
5. **Test changes** - Verify configuration changes work as expected