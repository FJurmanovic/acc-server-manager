package command

import "github.com/google/uuid"

type OutputCallback func(serverID uuid.UUID, output string, isError bool)

type CommandCallback func(serverID uuid.UUID, command string, args []string, completed bool, success bool, error string)

type CallbackConfig struct {
	OnOutput  OutputCallback
	OnCommand CommandCallback
}

func DefaultCallbackConfig() *CallbackConfig {
	return &CallbackConfig{
		OnOutput:  func(uuid.UUID, string, bool) {},
		OnCommand: func(uuid.UUID, string, []string, bool, bool, string) {},
	}
}
