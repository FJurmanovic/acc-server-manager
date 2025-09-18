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

type CallbackInteractiveCommandExecutor struct {
	*InteractiveCommandExecutor
	callbacks *CallbackConfig
	serverID  uuid.UUID
}

func NewCallbackInteractiveCommandExecutor(baseExecutor *CommandExecutor, tfaManager *model.Steam2FAManager, callbacks *CallbackConfig, serverID uuid.UUID) *CallbackInteractiveCommandExecutor {
	if callbacks == nil {
		callbacks = DefaultCallbackConfig()
	}

	return &CallbackInteractiveCommandExecutor{
		InteractiveCommandExecutor: &InteractiveCommandExecutor{
			CommandExecutor: baseExecutor,
			tfaManager:      tfaManager,
		},
		callbacks: callbacks,
		serverID:  serverID,
	}
}

func (e *CallbackInteractiveCommandExecutor) ExecuteInteractive(ctx context.Context, serverID *uuid.UUID, args ...string) error {
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

	logging.Info("Executing interactive command with callbacks: %s %s", e.ExePath, strings.Join(args, " "))

	e.callbacks.OnCommand(e.serverID, e.ExePath, args, false, false, "")
	e.callbacks.OnOutput(e.serverID, fmt.Sprintf("Starting command: %s %s", e.ExePath, strings.Join(args, " ")), false)

	if err := cmd.Start(); err != nil {
		e.callbacks.OnOutput(e.serverID, fmt.Sprintf("Failed to start command: %v", err), true)
		return fmt.Errorf("failed to start command: %v", err)
	}

	outputDone := make(chan error, 1)
	cmdDone := make(chan error, 1)

	go e.monitorOutputWithCallbacks(ctx, stdout, stderr, serverID, outputDone)

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
			e.callbacks.OnOutput(e.serverID, "Command execution completed", false)
		case outputErr = <-outputDone:
			completedCount++
			logging.Info("Output monitoring completed")
		case <-ctx.Done():
			e.callbacks.OnOutput(e.serverID, "Command execution cancelled", true)
			return ctx.Err()
		}
	}

	if outputErr != nil {
		logging.Warn("Output monitoring error: %v", outputErr)
		e.callbacks.OnOutput(e.serverID, fmt.Sprintf("Output monitoring error: %v", outputErr), true)
	}

	success := cmdErr == nil
	errorMsg := ""
	if cmdErr != nil {
		errorMsg = cmdErr.Error()
	}
	e.callbacks.OnCommand(e.serverID, e.ExePath, args, true, success, errorMsg)

	return cmdErr
}

func (e *CallbackInteractiveCommandExecutor) monitorOutputWithCallbacks(ctx context.Context, stdout, stderr io.Reader, serverID *uuid.UUID, done chan error) {
	defer func() {
		select {
		case done <- nil:
		default:
		}
	}()

	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)

	outputChan := make(chan outputLine, 100)
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
			e.callbacks.OnOutput(e.serverID, line, false)

			select {
			case outputChan <- outputLine{text: line, isError: false}:
			case <-ctx.Done():
				return
			}
		}
		if err := stdoutScanner.Err(); err != nil {
			logging.Warn("Stdout scanner error: %v", err)
			e.callbacks.OnOutput(e.serverID, fmt.Sprintf("Stdout scanner error: %v", err), true)
		}
	}()

	go func() {
		defer func() { readersDone <- struct{}{} }()
		for stderrScanner.Scan() {
			line := stderrScanner.Text()
			if e.LogOutput {
				logging.Info("STDERR: %s", line)
			}
			e.callbacks.OnOutput(e.serverID, line, true)

			select {
			case outputChan <- outputLine{text: line, isError: true}:
			case <-ctx.Done():
				return
			}
		}
		if err := stderrScanner.Err(); err != nil {
			logging.Warn("Stderr scanner error: %v", err)
			e.callbacks.OnOutput(e.serverID, fmt.Sprintf("Stderr scanner error: %v", err), true)
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
				for lineData := range outputChan {
					if e.is2FAPrompt(lineData.text) {
						if err := e.handle2FAPrompt(ctx, lineData.text, serverID); err != nil {
							logging.Error("Failed to handle 2FA prompt: %v", err)
							e.callbacks.OnOutput(e.serverID, fmt.Sprintf("Failed to handle 2FA prompt: %v", err), true)
							done <- err
							return
						}
					}
				}
				return
			}
		case lineData, ok := <-outputChan:
			if !ok {
				return
			}

			lowerLine := strings.ToLower(lineData.text)
			if strings.Contains(lowerLine, "steam console client") && strings.Contains(lowerLine, "valve corporation") {
				steamConsoleStarted = true
				logging.Info("Steam Console Client startup detected - will monitor for 2FA hang")
				e.callbacks.OnOutput(e.serverID, "Steam Console Client startup detected", false)
			}

			if e.is2FAPrompt(lineData.text) {
				if !tfaRequestCreated {
					e.callbacks.OnOutput(e.serverID, "2FA prompt detected - waiting for user confirmation", false)
					if err := e.handle2FAPrompt(ctx, lineData.text, serverID); err != nil {
						logging.Error("Failed to handle 2FA prompt: %v", err)
						e.callbacks.OnOutput(e.serverID, fmt.Sprintf("Failed to handle 2FA prompt: %v", err), true)
						done <- err
						return
					}
					tfaRequestCreated = true
				}
			}

			if tfaRequestCreated && e.isSteamContinuing(lineData.text) {
				logging.Info("Steam CMD appears to have continued after 2FA confirmation")
				e.callbacks.OnOutput(e.serverID, "Steam CMD continued after 2FA confirmation", false)
				e.autoCompletePendingRequests(serverID)
			}
		case <-time.After(15 * time.Second):
			if steamConsoleStarted && !tfaRequestCreated {
				logging.Info("Steam Console started but no output for 15 seconds - likely waiting for Steam Guard 2FA")
				e.callbacks.OnOutput(e.serverID, "Waiting for Steam Guard 2FA confirmation...", false)
				if err := e.handle2FAPrompt(ctx, "Steam CMD appears to be waiting for Steam Guard confirmation after startup", serverID); err != nil {
					logging.Error("Failed to handle Steam Guard 2FA prompt: %v", err)
					e.callbacks.OnOutput(e.serverID, fmt.Sprintf("Failed to handle Steam Guard 2FA prompt: %v", err), true)
					done <- err
					return
				}
				tfaRequestCreated = true
			}
		}
	}
}

type outputLine struct {
	text    string
	isError bool
}
