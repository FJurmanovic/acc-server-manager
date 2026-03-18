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
	"strings"
	"time"

	"github.com/google/uuid"
)

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

func (e *InteractiveCommandExecutor) ExecuteInteractive(ctx context.Context, serverID *uuid.UUID, args ...string) error {
	cmd := exec.CommandContext(ctx, e.ExePath, args...)

	if e.WorkDir != "" {
		cmd.Dir = e.WorkDir
	}

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

	debugMode := os.Getenv("STEAMCMD_DEBUG") == "true"
	if debugMode {
		logging.Info("STEAMCMD_DEBUG mode enabled - will log all output and create proactive 2FA requests")
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	outputDone := make(chan error, 1)
	cmdDone := make(chan error, 1)

	go e.monitorOutput(ctx, stdout, stderr, serverID, outputDone)

	go func() {
		cmdDone <- cmd.Wait()
	}()

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

	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)

	outputChan := make(chan string, 100)
	readersDone := make(chan struct{}, 2)

	steamConsoleStarted := false
	tfaRequestCreated := false

	go func() {
		defer func() { readersDone <- struct{}{} }()
		for stdoutScanner.Scan() {
			line := stdoutScanner.Text()
			if e.LogOutput {
				logging.Info("STDOUT: %s", line)
			}
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

	go func() {
		defer func() { readersDone <- struct{}{} }()
		for stderrScanner.Scan() {
			line := stderrScanner.Text()
			if e.LogOutput {
				logging.Info("STDERR: %s", line)
			}
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

	readersFinished := 0
	for {
		select {
		case <-ctx.Done():
			done <- ctx.Err()
			return
		case <-readersDone:
			readersFinished++
			if readersFinished == 2 {
				close(outputChan)
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
				return
			}

			lowerLine := strings.ToLower(line)
			if strings.Contains(lowerLine, "steam console client") && strings.Contains(lowerLine, "valve corporation") {
				steamConsoleStarted = true
				logging.Info("Steam Console Client startup detected - will monitor for 2FA hang")
			}

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

			if tfaRequestCreated && e.isSteamContinuing(line) {
				logging.Info("Steam CMD appears to have continued after 2FA confirmation - auto-completing 2FA request")
				e.autoCompletePendingRequests(serverID)
			}
		case <-time.After(15 * time.Second):
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

	request := e.tfaManager.CreateRequest(promptLine, serverID)
	logging.Info("Created 2FA request with ID: %s", request.ID)

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
