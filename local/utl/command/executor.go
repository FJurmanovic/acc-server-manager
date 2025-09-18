package command

import (
	"acc-server-manager/local/utl/logging"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type CommandExecutor struct {
	ExePath   string
	WorkDir   string
	LogOutput bool
}

type CommandBuilder struct {
	args []string
}

func NewCommandBuilder() *CommandBuilder {
	return &CommandBuilder{
		args: make([]string, 0),
	}
}

func (b *CommandBuilder) Add(arg string) *CommandBuilder {
	b.args = append(b.args, arg)
	return b
}

func (b *CommandBuilder) AddPair(key, value string) *CommandBuilder {
	b.args = append(b.args, key, value)
	return b
}

func (b *CommandBuilder) AddFlag(flag string, value interface{}) *CommandBuilder {
	b.args = append(b.args, fmt.Sprintf("%s=%v", flag, value))
	return b
}

func (b *CommandBuilder) Build() []string {
	return b.args
}

func (e *CommandExecutor) Execute(args ...string) error {
	cmd := exec.Command(e.ExePath, args...)

	if e.WorkDir != "" {
		cmd.Dir = e.WorkDir
	}

	if e.LogOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	logging.Info("Executing command: %s %s", e.ExePath, strings.Join(args, " "))
	return cmd.Run()
}

func (e *CommandExecutor) ExecuteWithBuilder(builder *CommandBuilder) error {
	return e.Execute(builder.Build()...)
}

func (e *CommandExecutor) ExecuteWithOutput(args ...string) (string, error) {
	cmd := exec.Command(e.ExePath, args...)

	if e.WorkDir != "" {
		cmd.Dir = e.WorkDir
	}

	logging.Info("Executing command: %s %s", e.ExePath, strings.Join(args, " "))
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (e *CommandExecutor) ExecuteWithEnv(env []string, args ...string) error {
	cmd := exec.Command(e.ExePath, args...)

	if e.WorkDir != "" {
		cmd.Dir = e.WorkDir
	}

	if e.LogOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	cmd.Env = append(os.Environ(), env...)

	logging.Info("Executing command: %s %s", e.ExePath, strings.Join(args, " "))
	return cmd.Run()
}
