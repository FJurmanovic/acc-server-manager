# API Reference

ACC Server Manager provides a RESTful API for managing ACC dedicated servers.

## Interactive Documentation

**Swagger UI is available at: http://localhost:3000/swagger/**

The Swagger UI provides:
- Interactive API testing
- Request/response examples
- Schema definitions
- Authentication testing

For detailed Swagger usage, see [SWAGGER.md](SWAGGER.md).

## Base URL

```
http://localhost:3000/api/v1
```

## Authentication

The API uses JWT (JSON Web Token) authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

### Getting a Token

```http
POST /auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "your-password"
}
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid",
    "username": "admin",
    "role": "admin"
  }
}
```

## Core Endpoints

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/login` | Login |
| POST | `/auth/register` | Register new user |
| GET | `/auth/me` | Get current user |
| POST | `/auth/refresh` | Refresh token |

### Server Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/servers` | List all servers |
| POST | `/servers` | Create new server |
| GET | `/servers/{id}` | Get server details |
| PUT | `/servers/{id}` | Update server |
| DELETE | `/servers/{id}` | Delete server |

### Server Operations

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/servers/{id}/service/start` | Start server |
| POST | `/servers/{id}/service/stop` | Stop server |
| POST | `/servers/{id}/service/restart` | Restart server |
| GET | `/servers/{id}/service/status` | Get server status |

### Configuration

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/servers/{id}/config/{file}` | Get config file |
| PUT | `/servers/{id}/config/{file}` | Update config file |

Available config files:
- `configuration.json`
- `settings.json`
- `event.json`
- `eventRules.json`
- `assistRules.json`

### System

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/system/health` | Health check

## Request Examples

### Create Server

```http
POST /api/v1/servers
Content-Type: application/json
Authorization: Bearer <token>

{
  "name": "My ACC Server",
  "track": "monza",
  "maxClients": 30
}
```

### Update Configuration

```http
PUT /api/v1/servers/123/config/settings.json
Content-Type: application/json
Authorization: Bearer <token>

{
  "serverName": "My Updated Server",
  "adminPassword": "secret",
  "trackMedalsRequirement": 0,
  "safetyRatingRequirement": -1
}
```

### Start Server

```http
POST /api/v1/servers/123/start
Authorization: Bearer <token>
```

## Response Formats

### Success Response

```json
{
  "success": true,
  "data": { ... }
}
```

### Error Response

```json
{
  "success": false,
  "error": "Error message",
  "code": "ERROR_CODE"
}
```

## Status Codes

- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `500` - Internal Server Error

## Rate Limiting

- 100 requests per minute per IP
- 1000 requests per hour per user

## Additional Resources

### Swagger Documentation

For comprehensive interactive API documentation:
- **Swagger UI**: http://localhost:3000/swagger/
- **OpenAPI Spec**: http://localhost:3000/swagger/doc.json
- **Usage Guide**: [SWAGGER.md](SWAGGER.md)

### Client Libraries

Generate client libraries using the OpenAPI spec:
```bash
# Download spec
curl http://localhost:3000/swagger/doc.json -o swagger.json

# Generate client (example for JavaScript)
openapi-generator generate -i swagger.json -g javascript -o ./client
```
