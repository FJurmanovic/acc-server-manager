package model

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Steam2FAStatus string

const (
	Steam2FAStatusIdle     Steam2FAStatus = "idle"
	Steam2FAStatusPending  Steam2FAStatus = "pending"
	Steam2FAStatusComplete Steam2FAStatus = "complete"
	Steam2FAStatusError    Steam2FAStatus = "error"
)

type Steam2FARequest struct {
	ID          string         `json:"id"`
	Status      Steam2FAStatus `json:"status"`
	Message     string         `json:"message"`
	RequestTime time.Time      `json:"requestTime"`
	CompletedAt *time.Time     `json:"completedAt,omitempty"`
	ErrorMsg    string         `json:"errorMsg,omitempty"`
	ServerID    *uuid.UUID     `json:"serverId,omitempty"`
}

// Steam2FAManager manages 2FA requests and responses
type Steam2FAManager struct {
	mu       sync.RWMutex
	requests map[string]*Steam2FARequest
	channels map[string]chan bool
}

func NewSteam2FAManager() *Steam2FAManager {
	return &Steam2FAManager{
		requests: make(map[string]*Steam2FARequest),
		channels: make(map[string]chan bool),
	}
}

func (m *Steam2FAManager) CreateRequest(message string, serverID *uuid.UUID) *Steam2FARequest {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := uuid.New().String()
	request := &Steam2FARequest{
		ID:          id,
		Status:      Steam2FAStatusPending,
		Message:     message,
		RequestTime: time.Now(),
		ServerID:    serverID,
	}

	m.requests[id] = request
	m.channels[id] = make(chan bool, 1)

	return request
}

func (m *Steam2FAManager) GetRequest(id string) (*Steam2FARequest, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	req, exists := m.requests[id]
	return req, exists
}

func (m *Steam2FAManager) GetPendingRequests() []*Steam2FARequest {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var pending []*Steam2FARequest
	for _, req := range m.requests {
		if req.Status == Steam2FAStatusPending {
			pending = append(pending, req)
		}
	}
	return pending
}

func (m *Steam2FAManager) CompleteRequest(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	req, exists := m.requests[id]
	if !exists {
		return fmt.Errorf("request %s not found", id)
	}

	if req.Status != Steam2FAStatusPending {
		return fmt.Errorf("request %s is not pending", id)
	}

	now := time.Now()
	req.Status = Steam2FAStatusComplete
	req.CompletedAt = &now

	// Signal the waiting goroutine
	if ch, exists := m.channels[id]; exists {
		select {
		case ch <- true:
		default:
		}
	}

	return nil
}

func (m *Steam2FAManager) ErrorRequest(id string, errorMsg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	req, exists := m.requests[id]
	if !exists {
		return fmt.Errorf("request %s not found", id)
	}

	req.Status = Steam2FAStatusError
	req.ErrorMsg = errorMsg

	// Signal the waiting goroutine with error
	if ch, exists := m.channels[id]; exists {
		select {
		case ch <- false:
		default:
		}
	}

	return nil
}

func (m *Steam2FAManager) WaitForCompletion(id string, timeout time.Duration) (bool, error) {
	m.mu.RLock()
	ch, exists := m.channels[id]
	m.mu.RUnlock()

	if !exists {
		return false, fmt.Errorf("request %s not found", id)
	}

	select {
	case success := <-ch:
		return success, nil
	case <-time.After(timeout):
		// Timeout - mark as error
		m.ErrorRequest(id, "timeout waiting for 2FA confirmation")
		return false, fmt.Errorf("timeout waiting for 2FA confirmation")
	}
}

func (m *Steam2FAManager) CleanupOldRequests(maxAge time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	for id, req := range m.requests {
		if req.RequestTime.Before(cutoff) {
			delete(m.requests, id)
			if ch, exists := m.channels[id]; exists {
				close(ch)
				delete(m.channels, id)
			}
		}
	}
}
