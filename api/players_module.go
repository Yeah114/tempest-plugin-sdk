package api

import "context"

type PlayersModule interface {
	Name() string

	NewPlayerKit(uuid string) PlayerKit
	GetAllOnlinePlayers(ctx context.Context) ([]PlayerKit, error)
	GetPlayerByName(ctx context.Context, name string) (PlayerKit, error)
	GetPlayerByUUID(ctx context.Context, uuid string) (PlayerKit, error)

	RegisterWhenPlayerChange(handler func(event *PlayerChangeEvent)) (string, error)
	UnregisterWhenPlayerChange(listenerID string) bool

	RawSayTo(target string, jsonText string) error
	SayTo(target string, message string) error
	RawTitleTo(target string, jsonText string) error
	TitleTo(target string, message string) error
	RawSubtitleTo(target string, subtitleJsonText, titleJsonText string) error
	SubtitleTo(target string, subtitleMessage, titleMessage string) error
	ActionBarTo(target string, message string) error
}

