# Logging System Usage Examples

This document provides comprehensive examples and documentation for using the new structured logging system in the ACC Server Manager.

## Overview

The logging system has been refactored to provide:
- **Structured logging** with separate files for each log level
- **Base logger** providing core functionality
- **Centralized error handling** for controllers
- **Backward compatibility** with existing code

## Architecture

```
logging/
├── base.go          # Base logger with core functionality
├── error.go         # Error-specific logging
├── warn.go          # Warning-specific logging
├── info.go          # Info-specific logging
├── debug.go         # Debug-specific logging
└── logger.go        # Main logger for backward compatibility
```

## Basic Usage

### Simple Logging (Backward Compatible)

```go
import "acc-server-manager/local/utl/logging"

// These work exactly as before
logging.Info("Server started on port %d", 8080)
logging.Error("Failed to connect to database: %v", err)
logging.Warn("Configuration value missing, using default")
logging.Debug("Processing request with ID: %s", requestID)
```

### Enhanced Logging with Context

```go
// Error logging with context
logging.ErrorWithContext("DATABASE", "Connection failed: %v", err)
logging.ErrorWithContext("AUTH", "Invalid credentials for user: %s", username)

// Info logging with context
logging.InfoWithContext("STARTUP", "Service initialized: %s", serviceName)
logging.InfoWithContext("SHUTDOWN", "Gracefully shutting down service: %s", serviceName)

// Warning with context
logging.WarnWithContext("CONFIG", "Missing configuration key: %s", configKey)

// Debug with context
logging.DebugWithContext("REQUEST", "Processing API call: %s", endpoint)
```

### Specialized Logging Functions

```go
// Application lifecycle
logging.LogStartup("DATABASE", "Connection pool initialized")
logging.LogShutdown("API_SERVER", "HTTP server stopped")

// Operations tracking
logging.LogOperation("USER_CREATE", "Created user with ID: " + userID.String())
logging.LogOperation("SERVER_START", "Started ACC server: " + serverName)

// HTTP request/response logging
logging.LogRequest("GET", "/api/v1/servers", "Mozilla/5.0...")
logging.LogResponse("GET", "/api/v1/servers", 200, "15ms")

// Error object logging
logging.LogError(err, "Failed to parse configuration file")

// Performance and debugging
logging.LogSQL("SELECT * FROM servers WHERE active = ?", true)
logging.LogMemory() // Logs current memory usage
logging.LogTiming("database_query", duration)
```

### Direct Logger Instances

```go
// Get specific logger instances for advanced usage
errorLogger := logging.GetErrorLogger()
infoLogger := logging.GetInfoLogger()
debugLogger := logging.GetDebugLogger()
warnLogger := logging.GetWarnLogger()

// Use specific logger methods
errorLogger.LogWithStackTrace("Critical system error occurred")
debugLogger.LogVariable("userConfig", userConfigObject)
debugLogger.LogState("cache", cacheState)
warnLogger.LogDeprecation("OldFunction", "NewFunction")
```

## Controller Error Handling

### Using the Centralized Error Handler

```go
package controller

import (
    "acc-server-manager/local/utl/error_handler"
    "github.com/gofiber/fiber/v2"
)

type MyController struct {
    service      *MyService
    errorHandler *error_handler.ControllerErrorHandler
}

func NewMyController(service *MyService) *MyController {
    return &MyController{
        service:      service,
        errorHandler: error_handler.NewControllerErrorHandler(),
    }
}

func (mc *MyController) GetUser(c *fiber.Ctx) error {
    userID, err := uuid.Parse(c.Params("id"))
    if err != nil {
        // Automatically logs error and returns standardized response
        return mc.errorHandler.HandleUUIDError(c, "user ID")
    }

    user, err := mc.service.GetUser(userID)
    if err != nil {
        // Logs error with context and returns appropriate HTTP status
        return mc.errorHandler.HandleServiceError(c, err)
    }

    return c.JSON(user)
}

func (mc *MyController) CreateUser(c *fiber.Ctx) error {
    var user User
    if err := c.BodyParser(&user); err != nil {
        return mc.errorHandler.HandleParsingError(c, err)
    }

    if err := mc.service.CreateUser(&user); err != nil {
        return mc.errorHandler.HandleServiceError(c, err)
    }

    return c.JSON(user)
}
```

### Available Error Handler Methods

```go
// Generic error handling
HandleError(c *fiber.Ctx, err error, statusCode int, context ...string)

// Specific error types
HandleValidationError(c *fiber.Ctx, err error, field string)
HandleDatabaseError(c *fiber.Ctx, err error)
HandleAuthError(c *fiber.Ctx, err error)
HandleNotFoundError(c *fiber.Ctx, resource string)
HandleBusinessLogicError(c *fiber.Ctx, err error)
HandleServiceError(c *fiber.Ctx, err error)
HandleParsingError(c *fiber.Ctx, err error)
HandleUUIDError(c *fiber.Ctx, field string)
```

### Global Error Handler Functions

```go
import "acc-server-manager/local/utl/error_handler"

// Use global error handler functions for convenience
func (mc *MyController) SomeEndpoint(c *fiber.Ctx) error {
    if err := someOperation(); err != nil {
        return error_handler.HandleServiceError(c, err)
    }
    return c.JSON(result)
}
```

## Request Logging Middleware

### Setup Request Logging

```go
import (
    middlewareLogging "acc-server-manager/local/middleware/logging"
)

func setupRoutes(app *fiber.App) {
    // Add request logging middleware
    app.Use(middlewareLogging.Handler())
    
    // Your routes here...
}
```

This will automatically log:
- Incoming requests with method, URL, and user agent
- Outgoing responses with status code and duration
- Any errors that occur during request processing

## Advanced Usage Examples

### Custom Logger with Specific Configuration

```go
// Create a custom base logger instance
baseLogger, err := logging.InitializeBase()
if err != nil {
    log.Fatal("Failed to initialize logger")
}

// Create specialized loggers
errorLogger := logging.NewErrorLogger()
debugLogger := logging.NewDebugLogger()

// Use them directly
errorLogger.LogWithStackTrace("Critical error in payment processing")
debugLogger.LogMemory()
debugLogger.LogGoroutines()
```

### Panic Recovery and Logging

```go
func dangerousOperation() {
    defer logging.RecoverAndLog()
    
    // Your potentially panicking code here
    // If panic occurs, it will be logged with full stack trace
}
```

### Performance Monitoring

```go
func processRequest() {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        logging.LogTiming("request_processing", duration)
    }()
    
    // Log memory usage periodically
    logging.LogMemory()
    
    // Your processing logic here
}
```

### Database Query Logging

```go
func (r *Repository) GetUser(id uuid.UUID) (*User, error) {
    query := "SELECT * FROM users WHERE id = ?"
    logging.LogSQL(query, id)
    
    var user User
    err := r.db.Get(&user, query, id)
    if err != nil {
        logging.ErrorWithContext("DATABASE", "Failed to get user %s: %v", id, err)
        return nil, err
    }
    
    return &user, nil
}
```

## Migration Guide

### For Existing Code

Your existing logging calls will continue to work:

```go
// These still work exactly as before
logging.Info("Message")
logging.Error("Error: %v", err)
logging.Warn("Warning message")
logging.Debug("Debug info")
```

### Upgrading to New Features

Consider upgrading to new features gradually:

```go
// Instead of:
logging.Error("Database error: %v", err)

// Use:
logging.ErrorWithContext("DATABASE", "Connection failed: %v", err)
// or
logging.LogError(err, "Database connection failed")
```

### Controller Updates

Replace manual error handling:

```go
// Old way:
if err != nil {
    logging.Error("Service error: %v", err)
    return c.Status(500).JSON(fiber.Map{"error": err.Error()})
}

// New way:
if err != nil {
    return mc.errorHandler.HandleServiceError(c, err)
}
```

## Configuration

### Log Levels

The system automatically handles different log levels. All logs are written to the same file but with different level indicators:

```
[2024-01-15 10:30:45.123] [INFO] [server.go:45] Server started successfully
[2024-01-15 10:30:46.456] [ERROR] [database.go:12] Connection failed: timeout
[2024-01-15 10:30:47.789] [WARN] [config.go:67] Using default configuration
[2024-01-15 10:30:48.012] [DEBUG] [handler.go:23] Processing request ID: 12345
```

### File Organization

Logs are automatically organized by date in the `logs/` directory:
- `logs/acc-server-2024-01-15.log`
- `logs/acc-server-2024-01-16.log`

### Output Destinations

All logs are written to both:
- Console (stdout) for development
- Log files for persistence

## Best Practices

1. **Use contextual logging** for better debugging
2. **Use appropriate log levels** (DEBUG for development, INFO for operations, WARN for issues, ERROR for failures)
3. **Use the error handler** in controllers for consistent error responses
4. **Include relevant information** in log messages (IDs, timestamps, etc.)
5. **Avoid logging sensitive information** (passwords, tokens, etc.)
6. **Use structured fields** when possible for better parsing

## Examples by Use Case

### API Controller Logging

```go
func (ac *APIController) CreateServer(c *fiber.Ctx) error {
    var server Server
    if err := c.BodyParser(&server); err != nil {
        return ac.errorHandler.HandleParsingError(c, err)
    }
    
    logging.InfoOperation("SERVER_CREATE", fmt.Sprintf("Creating server: %s", server.Name))
    
    if err := ac.service.CreateServer(&server); err != nil {
        return ac.errorHandler.HandleServiceError(c, err)
    }
    
    logging.InfoOperation("SERVER_CREATE", fmt.Sprintf("Successfully created server: %s (ID: %s)", server.Name, server.ID))
    return c.JSON(server)
}
```

### Service Layer Logging

```go
func (s *ServerService) StartServer(serverID uuid.UUID) error {
    logging.InfoWithContext("SERVER_SERVICE", "Starting server %s", serverID)
    
    server, err := s.repository.GetServer(serverID)
    if err != nil {
        logging.ErrorWithContext("SERVER_SERVICE", "Failed to get server %s: %v", serverID, err)
        return err
    }
    
    logging.DebugState("server_config", server)
    
    if err := s.processManager.Start(server); err != nil {
        logging.ErrorWithContext("SERVER_SERVICE", "Failed to start server %s: %v", serverID, err)
        return err
    }
    
    logging.InfoOperation("SERVER_START", fmt.Sprintf("Server %s started successfully", server.Name))
    return nil
}
```

This new logging system provides comprehensive error handling and logging capabilities while maintaining backward compatibility with existing code.