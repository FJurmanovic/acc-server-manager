package audit

import (
	"acc-server-manager/local/utl/logging"
	"context"
	"time"
)

type AuditAction string

const (
	ActionLogin          AuditAction = "LOGIN"
	ActionLogout         AuditAction = "LOGOUT"
	ActionServerCreate   AuditAction = "SERVER_CREATE"
	ActionServerUpdate   AuditAction = "SERVER_UPDATE"
	ActionServerDelete   AuditAction = "SERVER_DELETE"
	ActionServerStart    AuditAction = "SERVER_START"
	ActionServerStop     AuditAction = "SERVER_STOP"
	ActionUserCreate     AuditAction = "USER_CREATE"
	ActionUserUpdate     AuditAction = "USER_UPDATE"
	ActionUserDelete     AuditAction = "USER_DELETE"
	ActionConfigUpdate   AuditAction = "CONFIG_UPDATE"
	ActionSteamAuth      AuditAction = "STEAM_AUTH"
	ActionPermissionGrant AuditAction = "PERMISSION_GRANT"
	ActionPermissionRevoke AuditAction = "PERMISSION_REVOKE"
)

type AuditEntry struct {
	Timestamp time.Time   `json:"timestamp"`
	UserID    string      `json:"user_id"`
	Username  string      `json:"username"`
	Action    AuditAction `json:"action"`
	Resource  string      `json:"resource"`
	Details   string      `json:"details"`
	IPAddress string      `json:"ip_address"`
	UserAgent string      `json:"user_agent"`
	Success   bool        `json:"success"`
}

func LogAction(ctx context.Context, userID, username string, action AuditAction, resource, details, ipAddress, userAgent string, success bool) {
	logging.InfoWithContext("AUDIT", "User %s (%s) performed %s on %s from %s - Success: %t - Details: %s",
		username, userID, action, resource, ipAddress, success, details)
}

func LogAuthAction(ctx context.Context, username, ipAddress, userAgent string, success bool, details string) {
	action := ActionLogin
	if !success {
		details = "Failed: " + details
	}
	
	LogAction(ctx, "", username, action, "authentication", details, ipAddress, userAgent, success)
}

func LogServerAction(ctx context.Context, userID, username string, action AuditAction, serverID, ipAddress, userAgent string, success bool, details string) {
	LogAction(ctx, userID, username, action, "server:"+serverID, details, ipAddress, userAgent, success)
}

func LogUserManagementAction(ctx context.Context, adminUserID, adminUsername string, action AuditAction, targetUserID, ipAddress, userAgent string, success bool, details string) {
	LogAction(ctx, adminUserID, adminUsername, action, "user:"+targetUserID, details, ipAddress, userAgent, success)
}

func LogConfigAction(ctx context.Context, userID, username string, configType, ipAddress, userAgent string, success bool, details string) {
	LogAction(ctx, userID, username, ActionConfigUpdate, "config:"+configType, details, ipAddress, userAgent, success)
}