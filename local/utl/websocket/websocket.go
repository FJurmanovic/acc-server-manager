package websocket

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/utl/logging"
	"encoding/json"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

type WebSocketConnection struct {
	conn     *websocket.Conn
	serverID *uuid.UUID
	userID   *uuid.UUID
}

type WebSocketService struct {
	connections sync.Map
	mu          sync.RWMutex
}

func NewWebSocketService() *WebSocketService {
	return &WebSocketService{}
}

func (ws *WebSocketService) AddConnection(connID string, conn *websocket.Conn, userID *uuid.UUID) {
	wsConn := &WebSocketConnection{
		conn:   conn,
		userID: userID,
	}
	ws.connections.Store(connID, wsConn)
	logging.Info("WebSocket connection added: %s for user: %v", connID, userID)
}

func (ws *WebSocketService) RemoveConnection(connID string) {
	if conn, exists := ws.connections.LoadAndDelete(connID); exists {
		if wsConn, ok := conn.(*WebSocketConnection); ok {
			wsConn.conn.Close()
		}
	}
	logging.Info("WebSocket connection removed: %s", connID)
}

func (ws *WebSocketService) SetServerID(connID string, serverID uuid.UUID) {
	if conn, exists := ws.connections.Load(connID); exists {
		if wsConn, ok := conn.(*WebSocketConnection); ok {
			wsConn.serverID = &serverID
		}
	}
}

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

func (ws *WebSocketService) broadcastToServer(serverID uuid.UUID, message model.WebSocketMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		logging.Error("Failed to marshal WebSocket message: %v", err)
		return
	}

	ws.connections.Range(func(key, value interface{}) bool {
		if wsConn, ok := value.(*WebSocketConnection); ok {
			if wsConn.serverID != nil && *wsConn.serverID == serverID {
				if err := wsConn.conn.WriteMessage(websocket.TextMessage, data); err != nil {
					logging.Error("Failed to send WebSocket message to connection %s: %v", key, err)
					ws.RemoveConnection(key.(string))
				}
			}
		}
		return true
	})
}

func (ws *WebSocketService) BroadcastToUser(userID uuid.UUID, message model.WebSocketMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		logging.Error("Failed to marshal WebSocket message: %v", err)
		return
	}

	ws.connections.Range(func(key, value interface{}) bool {
		if wsConn, ok := value.(*WebSocketConnection); ok {
			if wsConn.userID != nil && *wsConn.userID == userID {
				if err := wsConn.conn.WriteMessage(websocket.TextMessage, data); err != nil {
					logging.Error("Failed to send WebSocket message to connection %s: %v", key, err)
					ws.RemoveConnection(key.(string))
				}
			}
		}
		return true
	})
}

func (ws *WebSocketService) GetActiveConnections() int {
	count := 0
	ws.connections.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}
