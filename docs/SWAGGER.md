# Swagger API Documentation

This guide explains how to use the interactive API documentation for ACC Server Manager.

## Accessing Swagger UI

The Swagger UI is available at:
```
http://localhost:3000/swagger/
```

## Authentication

Most API endpoints require JWT authentication. To use authenticated endpoints:

1. **Get a token**: Use the `/auth/login` endpoint with valid credentials
2. **Authorize**: Click the "Authorize" button in Swagger UI
3. **Enter token**: Type `Bearer <your-token>` (include the word "Bearer")
4. **Test endpoints**: Now you can test protected endpoints

## API Overview

### Authentication Endpoints
- `POST /auth/login` - Login and get JWT token
- `GET /auth/me` - Get current user information

### Server Management
- `GET /v1/server` - List all servers
- `POST /v1/server` - Create new server
- `GET /v1/server/{id}` - Get server details
- `PUT /v1/server/{id}` - Update server
- `DELETE /v1/server/{id}` - Delete server

### Server Configuration
- `GET /v1/server/{id}/config` - List config files
- `GET /v1/server/{id}/config/{file}` - Get config file
- `PUT /v1/server/{id}/config/{file}` - Update config file

### Service Control
- `GET /v1/service-control/{service}` - Get service status
- `POST /v1/service-control/start` - Start service
- `POST /v1/service-control/stop` - Stop service
- `POST /v1/service-control/restart` - Restart service

### Lookups
- `GET /v1/lookup/tracks` - Available tracks
- `GET /v1/lookup/car-models` - Available cars
- `GET /v1/lookup/driver-categories` - Driver categories
- `GET /v1/lookup/cup-categories` - Cup categories
- `GET /v1/lookup/session-types` - Session types

### User Management
- `GET /v1/membership` - List users
- `POST /v1/membership` - Create user
- `GET /v1/membership/{id}` - Get user details
- `PUT /v1/membership/{id}` - Update user
- `DELETE /v1/membership/{id}` - Delete user

## Common Operations

### Login Example
```json
POST /auth/login
{
  "username": "admin",
  "password": "your-password"
}

Response:
{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### Create Server Example
```json
POST /v1/server
Authorization: Bearer <token>
{
  "name": "My ACC Server",
  "track": "monza",
  "maxClients": 30,
  "tcpPort": 9201,
  "udpPort": 9201
}
```

### Update Configuration Example
```json
PUT /v1/server/{id}/config/settings.json
Authorization: Bearer <token>
{
  "serverName": "My Updated Server",
  "adminPassword": "secret",
  "trackMedalsRequirement": 0,
  "safetyRatingRequirement": -1
}
```

## Response Codes

- `200` - Success
- `201` - Created
- `400` - Bad Request (invalid input)
- `401` - Unauthorized (missing/invalid token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found
- `409` - Conflict (duplicate resource)
- `500` - Internal Server Error

## Testing Tips

1. **Start with login** - Get your token first
2. **Use "Try it out"** - Click this button to test endpoints
3. **Check examples** - Swagger shows request/response examples
4. **View schemas** - Click "Schema" to see data structures
5. **Download spec** - Get the OpenAPI spec at `/swagger/doc.json`

## Troubleshooting

### "Unauthorized" errors
- Ensure you've logged in and added the token
- Check token hasn't expired (24 hour expiry)
- Include "Bearer " prefix with token

### "Invalid JSON" errors
- Use the schema examples provided
- Validate JSON syntax
- Check required fields

### Can't see Swagger UI
- Ensure server is running
- Check correct URL and port
- Verify no firewall blocking

## Generating API Clients

You can generate client libraries from the OpenAPI spec:

1. Download spec from `/swagger/doc.json`
2. Use [OpenAPI Generator](https://openapi-generator.tech/)
3. Generate clients for your language:
   ```bash
   openapi-generator generate -i swagger.json -g javascript -o ./client
   ```

## API Rate Limits

- 100 requests per minute per IP
- 1000 requests per hour per user
- Rate limit headers included in responses

## For Developers

To update Swagger documentation:

1. Edit controller annotations
2. Run: `swag init -g cmd/api/swagger.go -o docs/`
3. Restart the server

See controller files for annotation examples.