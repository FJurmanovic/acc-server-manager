# Steam 2FA Implementation Documentation

## Overview

This document describes the implementation of Steam Two-Factor Authentication (2FA) support for the ACC Server Manager. When SteamCMD requires 2FA confirmation during server installation or updates, the system now signals the frontend and waits for user confirmation before proceeding.

## Architecture

The 2FA implementation consists of several interconnected components:

### Backend Components

1. **Steam2FAManager** (`local/model/steam_2fa.go`)
   - Thread-safe management of 2FA requests
   - Request lifecycle tracking (pending → complete/error)
   - Channel-based waiting mechanism for synchronization

2. **InteractiveCommandExecutor** (`local/utl/command/interactive_executor.go`)
   - Monitors SteamCMD output for 2FA prompts
   - Creates 2FA requests when prompts are detected
   - Waits for user confirmation before proceeding

3. **Steam2FAController** (`local/controller/steam_2fa.go`)
   - REST API endpoints for 2FA management
   - Handles frontend requests to complete/cancel 2FA

4. **Updated SteamService** (`local/service/steam_service.go`)
   - Uses InteractiveCommandExecutor for SteamCMD operations
   - Passes server context to 2FA requests

### Frontend Components

1. **Steam2FA Store** (`src/stores/steam2fa.ts`)
   - Svelte store for managing 2FA state
   - Automatic polling for pending requests
   - API communication methods

2. **Steam2FANotification Component** (`src/components/Steam2FANotification.svelte`)
   - Modal UI for 2FA confirmation
   - Automatic display when requests are pending
   - User interaction handling

3. **Type Definitions** (`src/models/steam2fa.ts`)
   - TypeScript interfaces for 2FA data structures

## API Endpoints

### GET /v1/steam2fa/pending
Returns all pending 2FA requests.

**Response:**
```json
[
  {
    "id": "uuid-string",
    "status": "pending",
    "message": "Steam Guard prompt message",
    "requestTime": "2024-01-01T12:00:00Z",
    "serverId": "server-uuid"
  }
]
```

### GET /v1/steam2fa/{id}
Returns a specific 2FA request by ID.

### POST /v1/steam2fa/{id}/complete
Marks a 2FA request as completed, allowing SteamCMD to proceed.

### POST /v1/steam2fa/{id}/cancel
Cancels a 2FA request, causing the SteamCMD operation to fail.

## Flow Diagram

```
SteamCMD Operation
        ↓
InteractiveCommandExecutor monitors output
        ↓
2FA prompt detected
        ↓
Steam2FARequest created
        ↓
Frontend polls and detects request
        ↓
Modal appears for user
        ↓
User confirms in Steam Mobile App
        ↓
User clicks "I've Confirmed"
        ↓
API call to complete request
        ↓
SteamCMD operation continues
```

## Configuration

### Backend Configuration

The system uses existing configuration patterns. No additional environment variables are required.

### Frontend Configuration

The API base URL is automatically configured as `/v1` to match the backend prefix.

Polling interval is set to 5 seconds by default and can be modified in `steam2fa.ts`:

```typescript
const POLLING_INTERVAL = 5000; // milliseconds
```

## Security Considerations

1. **Authentication Required**: All 2FA endpoints require user authentication
2. **Permission-Based Access**: Uses existing `ServerView` and `ServerUpdate` permissions
3. **Request Cleanup**: Automatic cleanup of old requests (30 minutes) prevents memory leaks
4. **No Sensitive Data**: No Steam credentials are exposed through the 2FA system

## Error Handling

### Backend Error Handling
- Timeouts after 5 minutes if no user response
- Proper error propagation to calling services
- Comprehensive logging for debugging

### Frontend Error Handling
- Network error handling with user feedback
- Automatic retry mechanisms
- Graceful degradation when API is unavailable

## Usage Instructions

### For Developers

1. **Adding New 2FA Prompts**: Extend the `is2FAPrompt` function in `interactive_executor.go`
2. **Customizing Timeouts**: Modify the timeout duration in `handle2FAPrompt`
3. **UI Customization**: Modify the `Steam2FANotification.svelte` component

### For Users

1. When creating or updating a server, watch for the 2FA notification
2. Check your Steam Mobile App when prompted
3. Confirm the login request in the Steam app
4. Click "I've Confirmed" in the web interface
5. The server operation will continue automatically

## Monitoring and Debugging

### Backend Logs
The system logs important events:
- 2FA prompt detection
- Request creation and completion
- Timeout events
- Error conditions

Search for log entries containing:
- `2FA prompt detected`
- `Created 2FA request`
- `2FA completed successfully`
- `2FA completion failed`

### Frontend Debugging
The Steam2FA store provides debugging information:
- `$steam2fa.error` - Current error state
- `$steam2fa.isLoading` - Loading state
- `$steam2fa.lastChecked` - Last polling timestamp

## Performance Considerations

1. **Polling Frequency**: 5-second polling provides good responsiveness without excessive load
2. **Request Cleanup**: Automatic cleanup prevents memory accumulation
3. **Efficient UI Updates**: Reactive Svelte stores minimize unnecessary re-renders

## Limitations

1. **Single User Sessions**: Currently designed for single-user scenarios
2. **Steam Mobile App Required**: Users must have Steam Mobile App installed
3. **Manual Confirmation**: No automatic 2FA code input support

## Future Enhancements

1. **WebSocket Support**: Real-time communication instead of polling
2. **Multiple User Support**: Handle multiple simultaneous 2FA requests
3. **Enhanced Prompt Detection**: More sophisticated Steam output parsing
4. **Notification System**: Browser notifications for 2FA requests

## Testing

### Manual Testing
1. Create a new server to trigger SteamCMD
2. Ensure Steam account has 2FA enabled
3. Verify modal appears when 2FA is required
4. Test both "confirm" and "cancel" workflows

### Automated Testing
The system includes comprehensive error handling but manual testing is recommended for 2FA workflows due to the interactive nature.

## Troubleshooting

### Common Issues

1. **Modal doesn't appear**
   - Check browser console for errors
   - Verify API connectivity
   - Ensure user has proper permissions

2. **SteamCMD hangs**
   - Check if 2FA request was created (backend logs)
   - Verify Steam Mobile App connectivity
   - Check for timeout errors

3. **API errors**
   - Verify user authentication
   - Check server permissions
   - Review backend error logs

### Debug Commands

```bash
# Check backend logs for 2FA events
grep -i "2fa" logs/app.log

# Monitor API requests
tail -f logs/app.log | grep "steam2fa"
```

## Version History

- **v1.0.0**: Initial implementation with polling-based frontend and REST API
- Added comprehensive error handling and logging
- Implemented automatic request cleanup
- Added responsive UI components

## Contributing

When contributing to the 2FA system:

1. Follow existing error handling patterns
2. Add comprehensive logging for new features
3. Update this documentation for any API changes
4. Test with actual Steam 2FA scenarios
5. Consider security implications of any changes