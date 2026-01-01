package api

import "time"

const NameCommandsModule = "commands"

type CommandOriginInfo struct {
	Origin    uint32
	UUID      string
	RequestID string
}

type CommandOutputMessage struct {
	Success    bool
	Message    string
	Parameters []string
}

type CommandOutput struct {
	CommandLine string

	Origin       CommandOriginInfo
	OutputType   byte
	SuccessCount uint32
	Messages     []CommandOutputMessage
	DataSet      string
}

type CommandsModule interface {
	Name() string

	SendSettingsCommand(command string, dimensional bool) error
	SendPlayerCommand(command string) error
	SendWSCommand(command string) error

	// SendPlayerCommandWithResp blocks until output arrives.
	// timeout <= 0 means no timeout (wait forever).
	SendPlayerCommandWithResp(command string, timeout time.Duration) (*CommandOutput, error)

	// SendWSCommandWithResp blocks until output arrives.
	// timeout <= 0 means no timeout (wait forever).
	SendWSCommandWithResp(command string, timeout time.Duration) (*CommandOutput, error)

	AwaitChangesGeneral() error
	SendChat(content string) error
	Title(message string) error
}

