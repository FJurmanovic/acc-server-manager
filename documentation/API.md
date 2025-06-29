# API Documentation for ACC Server Manager

## Overview

The ACC Server Manager provides a comprehensive REST API for managing Assetto Corsa Competizione dedicated servers. This API enables full control over server instances, configurations, user management, and monitoring through HTTP endpoints.

## Base URL

```
http://localhost:3000/api/v1
```

## Authentication

All API endpoints (except public ones) require authentication via JWT tokens.

### Authentication Header
```http
Authorization: Bearer <your-jwt-token>
```

### Token Expiration
- Default token lifetime: 24 hours
- Tokens should be refreshed before expiration
- Failed authentication returns HTTP 401

## Rate Limiting

The API implements multiple layers of rate limiting:

- **Global**: 100 requests per minute per IP
- **Authentication**: 5 attempts per 15 minutes per IP
- **API Endpoints**: 60 requests per minute per IP

Rate limit exceeded responses return HTTP 429 with retry information.

## Response Format

All API responses follow a consistent JSON format:

### Success Response
```json
{
  "success": true,
  "data": {
    // Response data
  },
  "message": "Operation completed successfully"
}
```

### Error Response
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": {}
  }
}
```

## HTTP Status Codes

| Status Code | Description |
|-------------|-------------|
| 200 | OK - Request successful |
| 201 | Created - Resource created successfully |
| 400 | Bad Request - Invalid request data |
| 401 | Unauthorized - Authentication required |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found - Resource not found |
| 409 | Conflict - Resource already exists |
| 422 | Unprocessable Entity - Validation failed |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error - Server error |

## API Endpoints

### Authentication

#### Login
```http
POST /api/v1/auth/login
```

**Request Body:**
```json
{
  "username": "string",
  "password": "string"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "token": "jwt-token-string",
    "user": {
      "id": "uuid",
      "username": "string",
      "role": {
        "id": "uuid",
        "name": "string",
        "permissions": []
      }
    }
  }
}
```

#### Register User
```http
POST /api/v1/auth/register
```
*Requires: `user.create` permission*

**Request Body:**
```json
{
  "username": "string",
  "password": "string",
  "roleId": "uuid"
}
```

#### Get Current User
```http
GET /api/v1/auth/me
```
*Requires: Authentication*

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "username": "string",
    "role": {
      "id": "uuid",
      "name": "string",
      "permissions": []
    }
  }
}
```

### Server Management

#### List Servers
```http
GET /api/v1/servers
```
*Requires: `server.read` permission*

**Query Parameters:**
- `page` (integer): Page number (default: 1)
- `limit` (integer): Items per page (default: 10)
- `search` (string): Search term
- `status` (string): Filter by status (running, stopped, error)

**Response:**
```json
{
  "success": true,
  "data": {
    "servers": [
      {
        "id": 1,
        "name": "string",
        "ip": "string",
        "port": 9600,
        "path": "string",
        "serviceName": "string",
        "status": "string",
        "dateCreated": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 50,
      "pages": 5
    }
  }
}
```

#### Create Server
```http
POST /api/v1/servers
```
*Requires: `server.create` permission*

**Request Body:**
```json
{
  "name": "string",
  "ip": "string",
  "port": 9600,
  "path": "string"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "string",
    "ip": "string",
    "port": 9600,
    "path": "string",
    "serviceName": "string",
    "status": "created",
    "dateCreated": "2024-01-01T00:00:00Z"
  }
}
```

#### Get Server Details
```http
GET /api/v1/servers/{id}
```
*Requires: `server.read` permission*

**Path Parameters:**
- `id` (integer): Server ID

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "string",
    "ip": "string",
    "port": 9600,
    "path": "string",
    "serviceName": "string",
    "status": "string",
    "dateCreated": "2024-01-01T00:00:00Z",
    "configs": [],
    "statistics": {}
  }
}
```

#### Update Server
```http
PUT /api/v1/servers/{id}
```
*Requires: `server.update` permission*

**Path Parameters:**
- `id` (integer): Server ID

**Request Body:**
```json
{
  "name": "string",
  "ip": "string",
  "port": 9600,
  "path": "string"
}
```

#### Delete Server
```http
DELETE /api/v1/servers/{id}
```
*Requires: `server.delete` permission*

**Path Parameters:**
- `id` (integer): Server ID

#### Start Server
```http
POST /api/v1/servers/{id}/start
```
*Requires: `server.control` permission*

#### Stop Server
```http
POST /api/v1/servers/{id}/stop
```
*Requires: `server.control` permission*

#### Restart Server
```http
POST /api/v1/servers/{id}/restart
```
*Requires: `server.control` permission*

### Configuration Management

#### Get Configuration File
```http
GET /api/v1/servers/{id}/config/{file}
```
*Requires: `config.read` permission*

**Path Parameters:**
- `id` (integer): Server ID
- `file` (string): Configuration file name (configuration, event, eventRules, settings)

**Response:**
```json
{
  "success": true,
  "data": {
    "file": "configuration",
    "content": {},
    "lastModified": "2024-01-01T00:00:00Z"
  }
}
```

#### Update Configuration File
```http
PUT /api/v1/servers/{id}/config/{file}
```
*Requires: `config.update` permission*

**Path Parameters:**
- `id` (integer): Server ID
- `file` (string): Configuration file name

**Query Parameters:**
- `restart` (boolean): Restart server after update (default: false)
- `override` (boolean): Override validation warnings (default: false)

**Request Body:**
```json
{
  "tcpPort": 9600,
  "udpPort": 9600,
  "maxConnections": 30,
  "registerToLobby": 1,
  "serverName": "My ACC Server",
  "password": "",
  "adminPassword": "admin123",
  "trackMedalsRequirement": 0,
  "safetyRatingRequirement": -1,
  "racecraftRatingRequirement": -1,
  "configVersion": 1
}
```

#### Validate Configuration
```http
POST /api/v1/servers/{id}/config/{file}/validate
```
*Requires: `config.read` permission*

**Request Body:** Configuration object to validate

**Response:**
```json
{
  "success": true,
  "data": {
    "valid": true,
    "errors": [],
    "warnings": []
  }
}
```

### Steam Integration

#### Get Steam Credentials
```http
GET /api/v1/steam/credentials
```
*Requires: `steam.read` permission*

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "username": "steam_username",
    "dateCreated": "2024-01-01T00:00:00Z",
    "lastUpdated": "2024-01-01T00:00:00Z"
  }
}
```

#### Update Steam Credentials
```http
PUT /api/v1/steam/credentials
```
*Requires: `steam.update` permission*

**Request Body:**
```json
{
  "username": "steam_username",
  "password": "steam_password"
}
```

#### Install/Update Server
```http
POST /api/v1/steam/install
```
*Requires: `steam.install` permission*

**Request Body:**
```json
{
  "serverId": 1,
  "validate": true,
  "beta": false
}
```

### User Management

#### List Users
```http
GET /api/v1/users
```
*Requires: `user.read` permission*

**Query Parameters:**
- `page` (integer): Page number
- `limit` (integer): Items per page
- `search` (string): Search term

**Response:**
```json
{
  "success": true,
  "data": {
    "users": [
      {
        "id": "uuid",
        "username": "string",
        "role": {
          "id": "uuid",
          "name": "string"
        },
        "createdAt": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {}
  }
}
```

#### Create User
```http
POST /api/v1/users
```
*Requires: `user.create` permission*

#### Update User
```http
PUT /api/v1/users/{id}
```
*Requires: `user.update` permission*

#### Delete User
```http
DELETE /api/v1/users/{id}
```
*Requires: `user.delete` permission*

### Role and Permission Management

#### List Roles
```http
GET /api/v1/roles
```
*Requires: `role.read` permission*

#### Create Role
```http
POST /api/v1/roles
```
*Requires: `role.create` permission*

**Request Body:**
```json
{
  "name": "string",
  "description": "string",
  "permissions": ["permission1", "permission2"]
}
```

#### List Permissions
```http
GET /api/v1/permissions
```
*Requires: `permission.read` permission*

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "name": "server.create",
      "description": "Create new servers",
      "category": "server"
    }
  ]
}
```

### Monitoring and Analytics

#### Get Server Statistics
```http
GET /api/v1/servers/{id}/stats
```
*Requires: `stats.read` permission*

**Query Parameters:**
- `from` (string): Start date (ISO 8601)
- `to` (string): End date (ISO 8601)
- `granularity` (string): hour, day, week, month

**Response:**
```json
{
  "success": true,
  "data": {
    "totalPlaytime": 3600,
    "playerCount": [],
    "sessionTypes": [],
    "dailyActivity": [],
    "recentSessions": []
  }
}
```

#### Get System Health
```http
GET /api/v1/system/health
```
*Public endpoint*

**Response:**
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "version": "1.0.0",
    "uptime": 3600,
    "database": "connected",
    "services": {
      "steam": "available",
      "nssm": "available"
    }
  }
}
```

### Lookup Data

#### Get Tracks
```http
GET /api/v1/lookup/tracks
```
*Public endpoint*

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "name": "monza",
      "uniquePitBoxes": 29,
      "privateServerSlots": 60
    }
  ]
}
```

#### Get Car Models
```http
GET /api/v1/lookup/cars
```
*Public endpoint*

#### Get Driver Categories
```http
GET /api/v1/lookup/driver-categories
```
*Public endpoint*

#### Get Cup Categories
```http
GET /api/v1/lookup/cup-categories
```
*Public endpoint*

#### Get Session Types
```http
GET /api/v1/lookup/session-types
```
*Public endpoint*

## Webhooks

The API supports webhook notifications for server events:

### Server Status Changes
```json
{
  "event": "server.status.changed",
  "serverId": 1,
  "serverName": "My Server",
  "oldStatus": "stopped",
  "newStatus": "running",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### Configuration Updates
```json
{
  "event": "server.config.updated",
  "serverId": 1,
  "serverName": "My Server",
  "configFile": "configuration",
  "userId": "uuid",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## Error Codes

| Code | Description |
|------|-------------|
| `AUTH_REQUIRED` | Authentication required |
| `AUTH_INVALID` | Invalid credentials |
| `AUTH_EXPIRED` | Token expired |
| `PERMISSION_DENIED` | Insufficient permissions |
| `VALIDATION_FAILED` | Request validation failed |
| `RESOURCE_NOT_FOUND` | Requested resource not found |
| `RESOURCE_EXISTS` | Resource already exists |
| `RATE_LIMIT_EXCEEDED` | Rate limit exceeded |
| `SERVER_ERROR` | Internal server error |
| `SERVICE_UNAVAILABLE` | External service unavailable |

## SDK Examples

### JavaScript/Node.js
```javascript
const axios = require('axios');

class ACCServerManagerAPI {
  constructor(baseUrl, token) {
    this.baseUrl = baseUrl;
    this.token = token;
    this.client = axios.create({
      baseURL: baseUrl,
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });
  }

  async getServers() {
    const response = await this.client.get('/servers');
    return response.data;
  }

  async createServer(serverData) {
    const response = await this.client.post('/servers', serverData);
    return response.data;
  }

  async updateConfig(serverId, file, config, restart = false) {
    const response = await this.client.put(
      `/servers/${serverId}/config/${file}?restart=${restart}`,
      config
    );
    return response.data;
  }
}

// Usage
const api = new ACCServerManagerAPI('http://localhost:3000/api/v1', 'your-jwt-token');
const servers = await api.getServers();
```

### Python
```python
import requests

class ACCServerManagerAPI:
    def __init__(self, base_url, token):
        self.base_url = base_url
        self.headers = {
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        }
    
    def get_servers(self):
        response = requests.get(f'{self.base_url}/servers', headers=self.headers)
        return response.json()
    
    def create_server(self, server_data):
        response = requests.post(
            f'{self.base_url}/servers',
            json=server_data,
            headers=self.headers
        )
        return response.json()

# Usage
api = ACCServerManagerAPI('http://localhost:3000/api/v1', 'your-jwt-token')
servers = api.get_servers()
```

### Go
```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type ACCServerManagerAPI struct {
    BaseURL string
    Token   string
    Client  *http.Client
}

func NewACCServerManagerAPI(baseURL, token string) *ACCServerManagerAPI {
    return &ACCServerManagerAPI{
        BaseURL: baseURL,
        Token:   token,
        Client:  &http.Client{},
    }
}

func (api *ACCServerManagerAPI) request(method, endpoint string, body interface{}) (*http.Response, error) {
    var reqBody bytes.Buffer
    if body != nil {
        json.NewEncoder(&reqBody).Encode(body)
    }
    
    req, err := http.NewRequest(method, api.BaseURL+endpoint, &reqBody)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", "Bearer "+api.Token)
    req.Header.Set("Content-Type", "application/json")
    
    return api.Client.Do(req)
}

func (api *ACCServerManagerAPI) GetServers() (interface{}, error) {
    resp, err := api.request("GET", "/servers", nil)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    return result, nil
}
```

## Best Practices

### Authentication
1. Store JWT tokens securely (httpOnly cookies for web apps)
2. Implement token refresh mechanism
3. Handle authentication errors gracefully
4. Use HTTPS in production

### Rate Limiting
1. Implement exponential backoff for rate-limited requests
2. Cache responses when appropriate
3. Use batch operations when available
4. Monitor rate limit headers

### Error Handling
1. Always check response status codes
2. Handle network errors gracefully
3. Implement retry logic for transient errors
4. Log errors for debugging

### Performance
1. Use pagination for large datasets
2. Implement client-side caching
3. Use WebSockets for real-time updates
4. Compress request/response bodies

## Support

For API support:
- **Documentation**: Check this guide and interactive Swagger UI
- **Issues**: Report API bugs via GitHub Issues
- **Community**: Join community discussions for help
- **Professional Support**: Contact maintainers for enterprise support

---

**Note**: This API is versioned. Breaking changes will result in a new API version. Always specify the version in your requests.