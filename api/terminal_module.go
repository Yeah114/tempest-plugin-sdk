package api

import (
	"context"
	"time"
)

const NameTerminalModule = "terminal"

type Level string

const (
	LevelSuccess Level = "SUCC"
	LevelInfo    Level = "INFO"
	LevelWarn    Level = "WARN"
	LevelError   Level = "ERRO"
)

type TerminalModule interface {
	Name() string

	Print(level Level, scope string, msg string)
	Info(scope string, msg string)
	Warn(scope string, msg string)
	Error(scope string, msg string)
	Success(scope string, msg string)
	Raw(msg string)

	ColorTransANSI(msg string) string

	SubscribeLines(ctx context.Context) (<-chan string, error)
	InterceptNextLine(ctx context.Context, timeout time.Duration, handler func(string)) (func(), error)
}
