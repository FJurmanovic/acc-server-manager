package command

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/utl/logging"
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
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

	// Read from stdout
	go func() {
		defer func() { readersDone <- struct{}{} }()
		for stdoutScanner.Scan() {
			line := stdoutScanner.Text()
			if e.LogOutput {
				logging.Info("STDOUT: %s", line)
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

			// Check if this line indicates a 2FA prompt
			if e.is2FAPrompt(line) {
				if err := e.handle2FAPrompt(ctx, line, serverID); err != nil {
					logging.Error("Failed to handle 2FA prompt: %v", err)
					done <- err
					return
				}
			}
		}
	}
}

func (e *InteractiveCommandExecutor) is2FAPrompt(line string) bool {
	// Common SteamCMD 2FA prompts
	twoFAKeywords := []string{
		"please enter your steam guard code",
		"steam guard",
		"two-factor",
		"authentication code",
		"please check your steam mobile app",
		"confirm in application",
	}

	lowerLine := strings.ToLower(line)
	for _, keyword := range twoFAKeywords {
		if strings.Contains(lowerLine, keyword) {
			return true
		}
	}
	return false
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
