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

	wsRoutes := routeGroups.WebSocket
	wsRoutes.Use("/", wsc.upgradeWebSocket)
	wsRoutes.Get("/", websocket.New(wsc.handleWebSocket))

	return wsc
}

func (wsc *WebSocketController) upgradeWebSocket(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
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

		claims, err := wsc.jwtHandler.ValidateToken(token)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid authentication token")
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid user ID in token")
		}

		c.Locals("userID", userID)
		c.Locals("username", claims.UserID)

		return c.Next()
	}

	return fiber.NewError(fiber.StatusUpgradeRequired, "WebSocket upgrade required")
}

func (wsc *WebSocketController) handleWebSocket(c *websocket.Conn) {
	connID := uuid.New().String()

	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		logging.Error("Failed to get user ID from WebSocket connection")
		c.Close()
		return
	}

	username, _ := c.Locals("username").(string)
	logging.Info("WebSocket connection established for user: %s (ID: %s)", username, userID.String())

	wsc.webSocketService.AddConnection(connID, c, &userID)

	defer func() {
		wsc.webSocketService.RemoveConnection(connID)
		logging.Info("WebSocket connection closed for user: %s", username)
	}()

	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logging.Error("WebSocket error for user %s: %v", username, err)
			}
			break
		}

		switch messageType {
		case websocket.TextMessage:
			wsc.handleTextMessage(connID, userID, message)
		case websocket.BinaryMessage:
			logging.Debug("Received binary message from user %s (not supported)", username)
		case websocket.PingMessage:
			if err := c.WriteMessage(websocket.PongMessage, nil); err != nil {
				logging.Error("Failed to send pong to user %s: %v", username, err)
				break
			}
		}
	}
}

func (wsc *WebSocketController) handleTextMessage(connID string, userID uuid.UUID, message []byte) {
	logging.Debug("Received WebSocket message from user %s: %s", userID.String(), string(message))

	messageStr := string(message)
	if len(messageStr) > 10 && messageStr[:9] == "server_id" {
		if serverIDStr := messageStr[10:]; len(serverIDStr) > 0 {
			if serverID, err := uuid.Parse(serverIDStr); err == nil {
				wsc.webSocketService.SetServerID(connID, serverID)
				logging.Info("Associated WebSocket connection %s with server %s", connID, serverID.String())
			}
		}
	}
}

func (wsc *WebSocketController) GetWebSocketUpgrade() fiber.Handler {
	return wsc.upgradeWebSocket
}

func (wsc *WebSocketController) GetWebSocketHandler() func(*websocket.Conn) {
	return wsc.handleWebSocket
}

func (wsc *WebSocketController) BroadcastServerCreationProgress(serverID uuid.UUID, step string, status string, message string) {
	logging.Info("Broadcasting server creation progress: %s - %s: %s", serverID.String(), step, status)
}
