package controller

import (
	"acc-server-manager/local/middleware"
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/common"
	"acc-server-manager/local/utl/jwt"
	"acc-server-manager/local/utl/logging"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

type WebSocketController struct {
	webSocketService *service.WebSocketService
	jwtHandler       *jwt.OpenJWTHandler
}

// NewWebSocketController initializes WebSocketController
func NewWebSocketController(
	wsService *service.WebSocketService,
	jwtHandler *jwt.OpenJWTHandler,
	routeGroups *common.RouteGroups,
	auth *middleware.AuthMiddleware,
) *WebSocketController {
	wsc := &WebSocketController{
		webSocketService: wsService,
		jwtHandler:       jwtHandler,
	}

	// WebSocket routes
	wsRoutes := routeGroups.WebSocket
	wsRoutes.Use("/", wsc.upgradeWebSocket)
	wsRoutes.Get("/", websocket.New(wsc.handleWebSocket))

	return wsc
}

// upgradeWebSocket middleware to upgrade HTTP to WebSocket and validate authentication
func (wsc *WebSocketController) upgradeWebSocket(c *fiber.Ctx) error {
	// Check if it's a WebSocket upgrade request
	if websocket.IsWebSocketUpgrade(c) {
		// Validate JWT token from query parameter or header
		token := c.Query("token")
		if token == "" {
			token = c.Get("Authorization")
			if token != "" && len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}
		}

		if token == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "Missing authentication token")
		}

		// Validate the token
		claims, err := wsc.jwtHandler.ValidateToken(token)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid authentication token")
		}

		// Parse UserID string to UUID
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid user ID in token")
		}

		// Store user info in context for use in WebSocket handler
		c.Locals("userID", userID)
		c.Locals("username", claims.UserID) // Use UserID as username for now

		return c.Next()
	}

	return fiber.NewError(fiber.StatusUpgradeRequired, "WebSocket upgrade required")
}

// handleWebSocket handles WebSocket connections
func (wsc *WebSocketController) handleWebSocket(c *websocket.Conn) {
	// Generate a unique connection ID
	connID := uuid.New().String()

	// Get user info from locals (set by middleware)
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		logging.Error("Failed to get user ID from WebSocket connection")
		c.Close()
		return
	}

	username, _ := c.Locals("username").(string)
	logging.Info("WebSocket connection established for user: %s (ID: %s)", username, userID.String())

	// Add the connection to the service
	wsc.webSocketService.AddConnection(connID, c, &userID)

	// Handle connection cleanup
	defer func() {
		wsc.webSocketService.RemoveConnection(connID)
		logging.Info("WebSocket connection closed for user: %s", username)
	}()

	// Handle incoming messages from the client
	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logging.Error("WebSocket error for user %s: %v", username, err)
			}
			break
		}

		// Handle different message types
		switch messageType {
		case websocket.TextMessage:
			wsc.handleTextMessage(connID, userID, message)
		case websocket.BinaryMessage:
			logging.Debug("Received binary message from user %s (not supported)", username)
		case websocket.PingMessage:
			// Respond with pong
			if err := c.WriteMessage(websocket.PongMessage, nil); err != nil {
				logging.Error("Failed to send pong to user %s: %v", username, err)
				break
			}
		}
	}
}

// handleTextMessage processes text messages from the client
func (wsc *WebSocketController) handleTextMessage(connID string, userID uuid.UUID, message []byte) {
	logging.Debug("Received WebSocket message from user %s: %s", userID.String(), string(message))

	// Parse the message to handle different types of client requests
	// For now, we'll just log it. In the future, you might want to handle:
	// - Subscription to specific server creation processes
	// - Client heartbeat/keepalive
	// - Request for status updates

	// Example: If the message contains a server ID, associate this connection with that server
	// This is a simple implementation - you might want to use proper JSON parsing
	messageStr := string(message)
	if len(messageStr) > 10 && messageStr[:9] == "server_id" {
		// Extract server ID from message like "server_id:uuid"
		if serverIDStr := messageStr[10:]; len(serverIDStr) > 0 {
			if serverID, err := uuid.Parse(serverIDStr); err == nil {
				wsc.webSocketService.SetServerID(connID, serverID)
				logging.Info("Associated WebSocket connection %s with server %s", connID, serverID.String())
			}
		}
	}
}

// GetWebSocketUpgrade returns the WebSocket upgrade handler for use in other controllers
func (wsc *WebSocketController) GetWebSocketUpgrade() fiber.Handler {
	return wsc.upgradeWebSocket
}

// GetWebSocketHandler returns the WebSocket connection handler for use in other controllers
func (wsc *WebSocketController) GetWebSocketHandler() func(*websocket.Conn) {
	return wsc.handleWebSocket
}

// BroadcastServerCreationProgress is a helper method for other services to broadcast progress
func (wsc *WebSocketController) BroadcastServerCreationProgress(serverID uuid.UUID, step string, status string, message string) {
	// This can be used by the ServerService during server creation
	logging.Info("Broadcasting server creation progress: %s - %s: %s", serverID.String(), step, status)
}
