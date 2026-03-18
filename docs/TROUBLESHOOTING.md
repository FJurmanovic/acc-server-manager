# Troubleshooting Guide

Common issues and solutions for ACC Server Manager.

## Installation Issues

### "go: command not found"

**Solution**: Install Go from https://golang.org/dl/ and add to PATH.

### "steamcmd.exe not found"

**Solution**: 
1. Download SteamCMD from https://steamcdn-a.akamaihd.net/client/installer/steamcmd.zip
2. Extract to `C:\steamcmd\`
3. Or set `STEAMCMD_PATH` environment variable to your location

### "nssm.exe not found"

**Solution**:
1. Download NSSM from https://nssm.cc/download
2. Place `nssm.exe` in application directory
3. Or set `NSSM_PATH` environment variable

## Startup Issues

### "JWT_SECRET environment variable is required"

**Solution**: Run the setup script:
```powershell
.\scripts\generate-secrets.ps1
```

### "Failed to connect database"

**Solution**:
1. Check write permissions in application directory
2. Delete `acc.db` if corrupted and restart
3. Ensure no other instance is running

### Port already in use

**Solution**:
1. Change port in `.env` file: `PORT=8080`
2. Or stop the application using port 3000

## Server Management Issues

### "Failed to create firewall rule"

**Solution**: Run ACC Server Manager as Administrator.

### ACC server won't start

**Solutions**:
1. Check ACC server logs in server directory
2. Verify ports are not in use
3. Ensure Steam credentials are correct
4. Check Windows Event Viewer

### "Steam authentication failed"

**Solutions**:
1. Verify Steam credentials in Settings
2. Check if Steam Guard is enabled
3. Try logging into Steam manually first

## Performance Issues

### High CPU usage

**Solutions**:
1. Reduce number of active servers
2. Check for runaway ACC server processes
3. Restart ACC Server Manager

### High memory usage

**Solutions**:
1. Check database size (should be < 1GB)
2. Restart application to clear caches
3. Reduce log retention

## Authentication Issues

### Can't login

**Solutions**:
1. Check username and password
2. Clear browser cookies
3. Check logs for specific errors

### "Token expired"

**Solution**: Login again to get a new token.

## Configuration Issues

### Changes not saving

**Solutions**:
1. Check file permissions
2. Ensure valid JSON format
3. Check logs for validation errors

### Can't edit server config

**Solutions**:
1. Stop the server first
2. Check user permissions
3. Verify file isn't locked

## Network Issues

### Can't connect to server

**Solutions**:
1. Check Windows Firewall rules
2. Verify port forwarding on router
3. Ensure server is actually running

### API requests failing

**Solutions**:
1. Check CORS settings if using custom frontend
2. Verify authentication token
3. Check API endpoint URL

## Logging & Debugging

### Enable debug logging

Add to `.env` file:
```
LOG_LEVEL=debug
```

### Log locations

- Application logs: `logs/app.log`
- Error logs: `logs/error.log`
- ACC server logs: In each server's directory

## Common Error Messages

### "Permission denied"
- Run as Administrator
- Check file/folder permissions

### "Invalid configuration"
- Validate JSON syntax
- Check required fields

### "Database locked"
- Close other instances
- Restart application

### "Service installation failed"
- Ensure NSSM is available
- Run as Administrator
- Check service name conflicts

## Getting Help

If these solutions don't work:

1. Check the logs in `logs/` directory
2. Search existing GitHub issues
3. Create a new issue with:
   - Error message
   - Steps to reproduce
   - System information
   - Relevant log entries