package api

import "github.com/google/uuid"

const NamePlayersModule = "players"

const (
	EventTypeWhenPlayerChange    = "players:when_player_change"
	PlayerChangeEventTypeExist   = "exist"
	PlayerChangeEventTypeOnline  = "online"
	PlayerChangeEventTypeOffline = "offline"
)

type PlayerChangeEvent struct {
	UUID      uuid.UUID
	Name      string
	EventType string
}
