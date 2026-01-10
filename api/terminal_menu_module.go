package api

// Terminal menu module name (must match host module name).
const NameTerminalMenuModule = "terminal_menu"

const (
	EventTypeAddMenuEntry   = "event_type_add_menu_entry"
	EventTypeTerminalCall   = "event_type_terminal_call"
	EventTypePopBackendMenu = "event_type_pop_backend_menu"
)

type TerminalMenuEntry struct {
	Triggers     []string
	ArgumentHint string
	Usage        string

	// OnTrigger is invoked when the entry is selected.
	// It is not serialisable; remote implementations will bridge it using RPC/broker callbacks.
	OnTrigger func([]string)
}

type TerminalMenuModule interface {
	Name() string

	RegisterTerminalMenuEntry(entry *TerminalMenuEntry) error
	RemoveTerminalMenuEntry(entry *TerminalMenuEntry) bool
	PublishTerminalCall(line string)
	PublishPopBackendMenu()

	RegisterWhenAddMenuEntry(handler func(*TerminalMenuEntry)) (string, error)
	UnregisterWhenAddMenuEntry(listenerID string) bool

	RegisterWhenTerminalCall(handler func(string)) (string, error)
	UnregisterWhenTerminalCall(listenerID string) bool

	RegisterWhenPopBackendMenu(handler func(struct{})) (string, error)
	UnregisterWhenPopBackendMenu(listenerID string) bool
}
