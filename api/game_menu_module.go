package api

import "context"

const NameGameMenuModule = "game_menu"

// GameMenuEntryInfo is a serialisable snapshot of a game-menu entry.
// It contains the auto-assigned EntryID used for triggering/removal.
type GameMenuEntryInfo struct {
	EntryID      string
	Triggers     []string
	ArgumentHint string
	Usage        string
}

type GameMenuEntry struct {
	Triggers     []string
	ArgumentHint string
	Usage        string

	// OnTrigger is invoked when the entry is selected in the in-game menu.
	// It is not serialisable; remote implementations will bridge it using RPC/broker callbacks.
	OnTrigger func(chat *ChatMsg)
}

type GameMenuModule interface {
	Name() string

	// RegisterMenuEntry registers a new entry and returns its auto-assigned EntryID.
	RegisterGameMenuEntry(entry *GameMenuEntry) (string, error)
	RemoveMenuEntry(entryID string)

	// SubscribeEntries returns a channel that receives a replay of existing entries followed by future updates.
	// The returned cancel function stops the subscription.
	SubscribeEntries(ctx context.Context) (<-chan *GameMenuEntryInfo, func(), error)

	// TriggerEntry triggers an entry by id.
	// This is used by the built-in game_menu.lua driver.
	TriggerEntry(entryID string, chat *ChatMsg)
}
