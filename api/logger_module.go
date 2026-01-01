package api

const NameLoggerModule = "logger"

// LoggerModule writes log lines to the host's log storage.
// It is separate from TerminalModule: terminal output is for UI, logger is for persistent logs.
type LoggerModule interface {
	Name() string

	Log(scope string, level Level, msg string)
	Info(scope, msg string)
	Warn(scope, msg string)
	Error(scope, msg string)
	Success(scope, msg string)
}

