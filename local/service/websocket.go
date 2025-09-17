package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/utl/logging"
	"encoding/json"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

// WebSocketConnection represents a single WebSocket connection
type WebSocketConnection struct {
	conn     *websocket.Conn
	serverID *uuid.UUID // If connected to a specific server creation process
	userID   *uuid.UUID // User who owns this connection
}

// WebSocketService manages WebSocket connections and message broadcasting
type WebSocketService struct {
	connections sync.Map // map[string]*WebSocketConnection - key is connection ID
	mu          sync.RWMutex
}

// NewWebSocketService creates a new WebSocket service
func NewWebSocketService() *WebSocketService {
	return &WebSocketService{}
}

// AddConnection adds a new WebSocket connection
func (ws *WebSocketService) AddConnection(connID string, conn *websocket.Conn, userID *uuid.UUID) {
	wsConn := &WebSocketConnection{
		conn:   conn,
		userID: userID,
	}
	ws.connections.Store(connID, wsConn)
	logging.Info("WebSocket connection added: %s for user: %v", connID, userID)
}

// RemoveConnection removes a WebSocket connection
func (ws *WebSocketService) RemoveConnection(connID string) {
	if conn, exists := ws.connections.LoadAndDelete(connID); exists {
		if wsConn, ok := conn.(*WebSocketConnection); ok {
			wsConn.conn.Close()
		}
	}
	logging.Info("WebSocket connection removed: %s", connID)
}

// SetServerID associates a connection with a specific server creation process
func (ws *WebSocketService) SetServerID(connID string, serverID uuid.UUID) {
	if conn, exists := ws.connections.Load(connID); exists {
		if wsConn, ok := conn.(*WebSocketConnection); ok {
			wsConn.serverID = &serverID
		}
	}
}

// BroadcastStep sends a step update to all connections associated with a server
func (ws *WebSocketService) BroadcastStep(serverID uuid.UUID, step model.ServerCreationStep, status model.StepStatus, message string, errorMsg string) {
	stepMsg := model.StepMessage{
		Step:    step,
		Status:  status,
		Message: message,
		Error:   errorMsg,
	}

	wsMsg := model.WebSocketMessage{
		Type:      model.MessageTypeStep,
		ServerID:  &serverID,
		Timestamp: time.Now().Unix(),
		Data:      stepMsg,
	}

	ws.broadcastToServer(serverID, wsMsg)
}

// BroadcastSteamOutput sends Steam command output to all connections associated with a server
func (ws *WebSocketService) BroadcastSteamOutput(serverID uuid.UUID, output string, isError bool) {
	steamMsg := model.SteamOutputMessage{
		Output:  output,
		IsError: isError,
	}

	wsMsg := model.WebSocketMessage{
		Type:      model.MessageTypeSteamOutput,
		ServerID:  &serverID,
		Timestamp: time.Now().Unix(),
		Data:      steamMsg,
	}

	ws.broadcastToServer(serverID, wsMsg)
}

// BroadcastError sends an error message to all connections associated with a server
func (ws *WebSocketService) BroadcastError(serverID uuid.UUID, error string, details string) {
	errorMsg := model.ErrorMessage{
		Error:   error,
		Details: details,
	}

	wsMsg := model.WebSocketMessage{
		Type:      model.MessageTypeError,
		ServerID:  &serverID,
		Timestamp: time.Now().Unix(),
		Data:      errorMsg,
	}

	ws.broadcastToServer(serverID, wsMsg)
}

// BroadcastComplete sends a completion message to all connections associated with a server
func (ws *WebSocketService) BroadcastComplete(serverID uuid.UUID, success bool, message string) {
	completeMsg := model.CompleteMessage{
		ServerID: serverID,
		Success:  success,
		Message:  message,
	}

	wsMsg := model.WebSocketMessage{
		Type:      model.MessageTypeComplete,
		ServerID:  &serverID,
		Timestamp: time.Now().Unix(),
		Data:      completeMsg,
	}

	ws.broadcastToServer(serverID, wsMsg)
}

// broadcastToServer sends a message to all connections associated with a specific server
func (ws *WebSocketService) broadcastToServer(serverID uuid.UUID, message model.WebSocketMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		logging.Error("Failed to marshal WebSocket message: %v", err)
		return
	}

	ws.connections.Range(func(key, value interface{}) bool {
		if wsConn, ok := value.(*WebSocketConnection); ok {
			// Send to connections associated with this server
			if wsConn.serverID != nil && *wsConn.serverID == serverID {
				if err := wsConn.conn.WriteMessage(websocket.TextMessage, data); err != nil {
					logging.Error("Failed to send WebSocket message to connection %s: %v", key, err)
					// Remove the connection if it's broken
					ws.RemoveConnection(key.(string))
				}
			}
		}
		return true
	})
}

// BroadcastToUser sends a message to all connections owned by a specific user
func (ws *WebSocketService) BroadcastToUser(userID uuid.UUID, message model.WebSocketMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		logging.Error("Failed to marshal WebSocket message: %v", err)
		return
	}

	ws.connections.Range(func(key, value interface{}) bool {
		if wsConn, ok := value.(*WebSocketConnection); ok {
			// Send to connections owned by this user
			if wsConn.userID != nil && *wsConn.userID == userID {
				if err := wsConn.conn.WriteMessage(websocket.TextMessage, data); err != nil {
					logging.Error("Failed to send WebSocket message to connection %s: %v", key, err)
					// Remove the connection if it's broken
					ws.RemoveConnection(key.(string))
				}
			}
		}
		return true
	})
}

// GetActiveConnections returns the count of active connections
func (ws *WebSocketService) GetActiveConnections() int {
	count := 0
	ws.connections.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}
