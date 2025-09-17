package command

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/utl/logging"
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
)

// InteractiveCommandExecutor extends CommandExecutor to handle interactive commands
type InteractiveCommandExecutor struct {
	*CommandExecutor
	tfaManager *model.Steam2FAManager
}

func NewInteractiveCommandExecutor(baseExecutor *CommandExecutor, tfaManager *model.Steam2FAManager) *InteractiveCommandExecutor {
	return &InteractiveCommandExecutor{
		CommandExecutor: baseExecutor,
		tfaManager:      tfaManager,
	}
}

// ExecuteInteractive runs a command that may require 2FA input
func (e *InteractiveCommandExecutor) ExecuteInteractive(ctx context.Context, serverID *uuid.UUID, args ...string) error {
	cmd := exec.CommandContext(ctx, e.ExePath, args...)

	if e.WorkDir != "" {
		cmd.Dir = e.WorkDir
	}

	// Create pipes for stdin, stdout, and stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %v", err)
	}
	defer stdin.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}
	defer stdout.Close()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}
	defer stderr.Close()

	logging.Info("Executing interactive command: %s %s", e.ExePath, strings.Join(args, " "))

	// Enable debug mode if environment variable is set
	debugMode := os.Getenv("STEAMCMD_DEBUG") == "true"
	if debugMode {
		logging.Info("STEAMCMD_DEBUG mode enabled - will log all output and create proactive 2FA requests")
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	// Create channels for output monitoring
	outputDone := make(chan error, 1)
	cmdDone := make(chan error, 1)

	// Monitor stdout and stderr for 2FA prompts
	go e.monitorOutput(ctx, stdout, stderr, serverID, outputDone)

	// Wait for the command to finish in a separate goroutine
	go func() {
		cmdDone <- cmd.Wait()
	}()

	// Wait for both command and output monitoring to complete
	var cmdErr, outputErr error
	completedCount := 0

	for completedCount < 2 {
		select {
		case cmdErr = <-cmdDone:
			completedCount++
			logging.Info("Command execution completed")
		case outputErr = <-outputDone:
			completedCount++
			logging.Info("Output monitoring completed")
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	if outputErr != nil {
		logging.Warn("Output monitoring error: %v", outputErr)
	}

	return cmdErr
}

func (e *InteractiveCommandExecutor) monitorOutput(ctx context.Context, stdout, stderr io.Reader, serverID *uuid.UUID, done chan error) {
	defer func() {
		select {
		case done <- nil:
		default:
		}
	}()

	// Create scanners for both outputs
	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)

	outputChan := make(chan string, 100) // Buffered channel to prevent blocking
	readersDone := make(chan struct{}, 2)

	// Track Steam Console startup for this specific execution
	steamConsoleStarted := false
	tfaRequestCreated := false

	// Read from stdout
	go func() {
		defer func() { readersDone <- struct{}{} }()
		for stdoutScanner.Scan() {
			line := stdoutScanner.Text()
			if e.LogOutput {
				logging.Info("STDOUT: %s", line)
			}
			// Always log Steam CMD output for debugging 2FA issues
			if strings.Contains(strings.ToLower(line), "steam") {
				logging.Info("STEAM_DEBUG: %s", line)
			}
			select {
			case outputChan <- line:
			case <-ctx.Done():
				return
			}
		}
		if err := stdoutScanner.Err(); err != nil {
			logging.Warn("Stdout scanner error: %v", err)
		}
	}()

	// Read from stderr
	go func() {
		defer func() { readersDone <- struct{}{} }()
		for stderrScanner.Scan() {
			line := stderrScanner.Text()
			if e.LogOutput {
				logging.Info("STDERR: %s", line)
			}
			// Always log Steam CMD errors for debugging 2FA issues
			if strings.Contains(strings.ToLower(line), "steam") {
				logging.Info("STEAM_DEBUG_ERR: %s", line)
			}
			select {
			case outputChan <- line:
			case <-ctx.Done():
				return
			}
		}
		if err := stderrScanner.Err(); err != nil {
			logging.Warn("Stderr scanner error: %v", err)
		}
	}()

	// Monitor for completion and 2FA prompts
	readersFinished := 0
	for {
		select {
		case <-ctx.Done():
			done <- ctx.Err()
			return
		case <-readersDone:
			readersFinished++
			if readersFinished == 2 {
				// Both readers are done, close output channel and finish monitoring
				close(outputChan)
				// Drain any remaining output
				for line := range outputChan {
					if e.is2FAPrompt(line) {
						if err := e.handle2FAPrompt(ctx, line, serverID); err != nil {
							logging.Error("Failed to handle 2FA prompt: %v", err)
							done <- err
							return
						}
					}
				}
				return
			}
		case line, ok := <-outputChan:
			if !ok {
				// Channel closed, we're done
				return
			}

			// Check for Steam Console startup
			lowerLine := strings.ToLower(line)
			if strings.Contains(lowerLine, "steam console client") && strings.Contains(lowerLine, "valve corporation") {
				steamConsoleStarted = true
				logging.Info("Steam Console Client startup detected - will monitor for 2FA hang")
			}

			// Check if this line indicates a 2FA prompt
			if e.is2FAPrompt(line) {
				if !tfaRequestCreated {
					if err := e.handle2FAPrompt(ctx, line, serverID); err != nil {
						logging.Error("Failed to handle 2FA prompt: %v", err)
						done <- err
						return
					}
					tfaRequestCreated = true
				}
			}

			// Check if Steam CMD continued after 2FA (auto-completion)
			if tfaRequestCreated && e.isSteamContinuing(line) {
				logging.Info("Steam CMD appears to have continued after 2FA confirmation - auto-completing 2FA request")
				// Auto-complete any pending 2FA requests for this server
				e.autoCompletePendingRequests(serverID)
			}
		case <-time.After(15 * time.Second):
			// If Steam Console has started and we haven't seen output for 15 seconds,
			// it's very likely waiting for 2FA confirmation
			if steamConsoleStarted && !tfaRequestCreated {
				logging.Info("Steam Console started but no output for 15 seconds - likely waiting for Steam Guard 2FA")
				if err := e.handle2FAPrompt(ctx, "Steam CMD appears to be waiting for Steam Guard confirmation after startup", serverID); err != nil {
					logging.Error("Failed to handle Steam Guard 2FA prompt: %v", err)
					done <- err
					return
				}
				tfaRequestCreated = true
			} else if !steamConsoleStarted {
				logging.Info("No output for 15 seconds (Steam Console not yet started)")
			}
		}
	}
}

func (e *InteractiveCommandExecutor) is2FAPrompt(line string) bool {
	// Common SteamCMD 2FA prompts - updated with more comprehensive patterns
	twoFAKeywords := []string{
		"please enter your steam guard code",
		"steam guard",
		"two-factor",
		"authentication code",
		"please check your steam mobile app",
		"confirm in application",
		"enter the current code from your steam mobile app",
		"steam guard mobile authenticator",
		"waiting for user info",
		"login failure",
		"two factor code required",
		"enter steam guard code",
		"mobile authenticator code",
		"authenticator app",
		"guard code",
		"mobile app",
		"confirmation required",
	}

	lowerLine := strings.ToLower(line)
	for _, keyword := range twoFAKeywords {
		if strings.Contains(lowerLine, keyword) {
			logging.Info("2FA keyword match found: '%s' in line: '%s'", keyword, line)
			return true
		}
	}

	// Also check for patterns that might indicate Steam is waiting for input
	waitingPatterns := []string{
		"waiting for",
		"please enter",
		"enter code",
		"code:",
		"authenticator:",
	}

	for _, pattern := range waitingPatterns {
		if strings.Contains(lowerLine, pattern) {
			logging.Info("Potential 2FA waiting pattern found: '%s' in line: '%s'", pattern, line)
			return true
		}
	}

	return false
}

func (e *InteractiveCommandExecutor) isSteamContinuing(line string) bool {
	lowerLine := strings.ToLower(line)
	continuingPatterns := []string{
		"loading steam api",
		"logging in user",
		"waiting for client config",
		"waiting for user info",
		"update state",
		"success! app",
		"fully installed",
	}

	for _, pattern := range continuingPatterns {
		if strings.Contains(lowerLine, pattern) {
			return true
		}
	}
	return false
}

func (e *InteractiveCommandExecutor) autoCompletePendingRequests(serverID *uuid.UUID) {
	if e.tfaManager == nil {
		return
	}

	pendingRequests := e.tfaManager.GetPendingRequests()
	for _, req := range pendingRequests {
		if req.ServerID != nil && serverID != nil && *req.ServerID == *serverID {
			logging.Info("Auto-completing 2FA request %s for server %s", req.ID, serverID.String())
			if err := e.tfaManager.CompleteRequest(req.ID); err != nil {
				logging.Warn("Failed to auto-complete 2FA request %s: %v", req.ID, err)
			}
		}
	}
}

func (e *InteractiveCommandExecutor) handle2FAPrompt(_ context.Context, promptLine string, serverID *uuid.UUID) error {
	logging.Info("2FA prompt detected: %s", promptLine)

	// Create a 2FA request
	request := e.tfaManager.CreateRequest(promptLine, serverID)
	logging.Info("Created 2FA request with ID: %s", request.ID)

	// Wait for user to complete the 2FA process
	// Use a reasonable timeout (e.g., 5 minutes)
	timeout := 5 * time.Minute
	success, err := e.tfaManager.WaitForCompletion(request.ID, timeout)

	if err != nil {
		logging.Error("2FA completion failed: %v", err)
		return err
	}

	if !success {
		logging.Error("2FA was not completed successfully")
		return fmt.Errorf("2FA authentication failed")
	}

	logging.Info("2FA completed successfully")
	return nil
}

// WebSocketInteractiveCommandExecutor extends InteractiveCommandExecutor to stream output via WebSocket
type WebSocketInteractiveCommandExecutor struct {
	*InteractiveCommandExecutor
	wsService interface{} // Using interface{} to avoid circular import
	serverID  uuid.UUID
}

// NewInteractiveCommandExecutorWithWebSocket creates a new WebSocket-enabled interactive command executor
func NewInteractiveCommandExecutorWithWebSocket(baseExecutor *CommandExecutor, tfaManager *model.Steam2FAManager, wsService interface{}, serverID uuid.UUID) *WebSocketInteractiveCommandExecutor {
	return &WebSocketInteractiveCommandExecutor{
		InteractiveCommandExecutor: &InteractiveCommandExecutor{
			CommandExecutor: baseExecutor,
			tfaManager:      tfaManager,
		},
		wsService: wsService,
		serverID:  serverID,
	}
}

// ExecuteInteractive runs a command with WebSocket output streaming
func (e *WebSocketInteractiveCommandExecutor) ExecuteInteractive(ctx context.Context, serverID *uuid.UUID, args ...string) error {
	cmd := exec.CommandContext(ctx, e.ExePath, args...)

	if e.WorkDir != "" {
		cmd.Dir = e.WorkDir
	}

	// Create pipes for stdin, stdout, and stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %v", err)
	}
	defer stdin.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}
	defer stdout.Close()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}
	defer stderr.Close()

	logging.Info("Executing interactive command with WebSocket streaming: %s %s", e.ExePath, strings.Join(args, " "))

	// Broadcast command start via WebSocket
	e.broadcastSteamOutput(fmt.Sprintf("Starting command: %s %s", e.ExePath, strings.Join(args, " ")), false)

	if err := cmd.Start(); err != nil {
		e.broadcastSteamOutput(fmt.Sprintf("Failed to start command: %v", err), true)
		return fmt.Errorf("failed to start command: %v", err)
	}

	// Create channels for output monitoring
	outputDone := make(chan error, 1)
	cmdDone := make(chan error, 1)

	// Monitor stdout and stderr for 2FA prompts with WebSocket streaming
	go e.monitorOutputWithWebSocket(ctx, stdout, stderr, serverID, outputDone)

	// Wait for the command to finish in a separate goroutine
	go func() {
		cmdDone <- cmd.Wait()
	}()

	// Wait for both command and output monitoring to complete
	var cmdErr, outputErr error
	completedCount := 0

	for completedCount < 2 {
		select {
		case cmdErr = <-cmdDone:
			completedCount++
			logging.Info("Command execution completed")
			e.broadcastSteamOutput("Command execution completed", false)
		case outputErr = <-outputDone:
			completedCount++
			logging.Info("Output monitoring completed")
		case <-ctx.Done():
			e.broadcastSteamOutput("Command execution cancelled", true)
			return ctx.Err()
		}
	}

	if outputErr != nil {
		logging.Warn("Output monitoring error: %v", outputErr)
		e.broadcastSteamOutput(fmt.Sprintf("Output monitoring error: %v", outputErr), true)
	}

	return cmdErr
}

// broadcastSteamOutput sends output to WebSocket using reflection to avoid circular imports
func (e *WebSocketInteractiveCommandExecutor) broadcastSteamOutput(output string, isError bool) {
	if e.wsService == nil {
		return
	}

	// Use reflection to call BroadcastSteamOutput method
	wsServiceVal := reflect.ValueOf(e.wsService)
	method := wsServiceVal.MethodByName("BroadcastSteamOutput")
	if !method.IsValid() {
		logging.Warn("BroadcastSteamOutput method not found on WebSocket service")
		return
	}

	// Call the method with parameters: serverID, output, isError
	args := []reflect.Value{
		reflect.ValueOf(e.serverID),
		reflect.ValueOf(output),
		reflect.ValueOf(isError),
	}
	method.Call(args)
}

// monitorOutputWithWebSocket monitors command output and streams it via WebSocket
func (e *WebSocketInteractiveCommandExecutor) monitorOutputWithWebSocket(ctx context.Context, stdout, stderr io.Reader, serverID *uuid.UUID, done chan error) {
	defer func() {
		select {
		case done <- nil:
		default:
		}
	}()

	// Create scanners for both outputs
	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)

	outputChan := make(chan outputLine, 100) // Buffered channel to prevent blocking
	readersDone := make(chan struct{}, 2)

	// Track Steam Console startup for this specific execution
	steamConsoleStarted := false
	tfaRequestCreated := false

	// Read from stdout
	go func() {
		defer func() { readersDone <- struct{}{} }()
		for stdoutScanner.Scan() {
			line := stdoutScanner.Text()
			if e.LogOutput {
				logging.Info("STDOUT: %s", line)
			}
			// Stream output via WebSocket
			e.broadcastSteamOutput(line, false)

			select {
			case outputChan <- outputLine{text: line, isError: false}:
			case <-ctx.Done():
				return
			}
		}
		if err := stdoutScanner.Err(); err != nil {
			logging.Warn("Stdout scanner error: %v", err)
			e.broadcastSteamOutput(fmt.Sprintf("Stdout scanner error: %v", err), true)
		}
	}()

	// Read from stderr
	go func() {
		defer func() { readersDone <- struct{}{} }()
		for stderrScanner.Scan() {
			line := stderrScanner.Text()
			if e.LogOutput {
				logging.Info("STDERR: %s", line)
			}
			// Stream error output via WebSocket
			e.broadcastSteamOutput(line, true)

			select {
			case outputChan <- outputLine{text: line, isError: true}:
			case <-ctx.Done():
				return
			}
		}
		if err := stderrScanner.Err(); err != nil {
			logging.Warn("Stderr scanner error: %v", err)
			e.broadcastSteamOutput(fmt.Sprintf("Stderr scanner error: %v", err), true)
		}
	}()

	// Monitor for completion and 2FA prompts
	readersFinished := 0
	for {
		select {
		case <-ctx.Done():
			done <- ctx.Err()
			return
		case <-readersDone:
			readersFinished++
			if readersFinished == 2 {
				// Both readers are done, close output channel and finish monitoring
				close(outputChan)
				// Drain any remaining output
				for lineData := range outputChan {
					if e.is2FAPrompt(lineData.text) {
						if err := e.handle2FAPrompt(ctx, lineData.text, serverID); err != nil {
							logging.Error("Failed to handle 2FA prompt: %v", err)
							e.broadcastSteamOutput(fmt.Sprintf("Failed to handle 2FA prompt: %v", err), true)
							done <- err
							return
						}
					}
				}
				return
			}
		case lineData, ok := <-outputChan:
			if !ok {
				// Channel closed, we're done
				return
			}

			// Check for Steam Console startup
			lowerLine := strings.ToLower(lineData.text)
			if strings.Contains(lowerLine, "steam console client") && strings.Contains(lowerLine, "valve corporation") {
				steamConsoleStarted = true
				logging.Info("Steam Console Client startup detected - will monitor for 2FA hang")
				e.broadcastSteamOutput("Steam Console Client startup detected", false)
			}

			// Check if this line indicates a 2FA prompt
			if e.is2FAPrompt(lineData.text) {
				if !tfaRequestCreated {
					e.broadcastSteamOutput("2FA prompt detected - waiting for user confirmation", false)
					if err := e.handle2FAPrompt(ctx, lineData.text, serverID); err != nil {
						logging.Error("Failed to handle 2FA prompt: %v", err)
						e.broadcastSteamOutput(fmt.Sprintf("Failed to handle 2FA prompt: %v", err), true)
						done <- err
						return
					}
					tfaRequestCreated = true
				}
			}

			// Check if Steam CMD continued after 2FA (auto-completion)
			if tfaRequestCreated && e.isSteamContinuing(lineData.text) {
				logging.Info("Steam CMD appears to have continued after 2FA confirmation")
				e.broadcastSteamOutput("Steam CMD continued after 2FA confirmation", false)
				// Auto-complete any pending 2FA requests for this server
				e.autoCompletePendingRequests(serverID)
			}
		case <-time.After(15 * time.Second):
			// If Steam Console has started and we haven't seen output for 15 seconds,
			// it's very likely waiting for 2FA confirmation
			if steamConsoleStarted && !tfaRequestCreated {
				logging.Info("Steam Console started but no output for 15 seconds - likely waiting for Steam Guard 2FA")
				e.broadcastSteamOutput("Waiting for Steam Guard 2FA confirmation...", false)
				if err := e.handle2FAPrompt(ctx, "Steam CMD appears to be waiting for Steam Guard confirmation after startup", serverID); err != nil {
					logging.Error("Failed to handle Steam Guard 2FA prompt: %v", err)
					e.broadcastSteamOutput(fmt.Sprintf("Failed to handle Steam Guard 2FA prompt: %v", err), true)
					done <- err
					return
				}
				tfaRequestCreated = true
			}
		}
	}
}

// outputLine represents a line of output with error status
type outputLine struct {
	text    string
	isError bool
}
