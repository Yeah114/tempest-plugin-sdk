package api

import "context"

const NameGameMenuModule = "game_menu"

type GameMenuEntry struct {
	ID           string
	Triggers     []string
	ArgumentHint string
	Usage        string
}

type GameMenuModule interface {
	Name() string

	RegisterMenuEntry(entry *GameMenuEntry) error
	RemoveMenuEntry(id string)

	// SubscribeEntries returns a channel that receives a replay of existing entries followed by future updates.
	// The returned cancel function stops the subscription.
	SubscribeEntries(ctx context.Context) (<-chan *GameMenuEntry, func(), error)

	RegisterTriggerHandler(id string, handler func(chatJSON []byte)) error
	RemoveTriggerHandler(id string)
	TriggerEntry(id string, chatJSON []byte)
}

