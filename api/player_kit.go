package api

import (
	"context"
	"time"
)

// PlayerKit is a convenient wrapper for a single player.
// Implementations may represent offline players as well (as long as UUID/name is known).
type PlayerKit interface {
	GetUUIDString() string
	GetName() string

	GetEntityUniqueID(ctx context.Context) (int64, error)
	GetLoginTime(ctx context.Context) (time.Time, error)
	GetPlatformChatID(ctx context.Context) (string, error)
	GetBuildPlatform(ctx context.Context) (int32, error)
	GetSkinID(ctx context.Context) (string, error)

	GetCanBuild(ctx context.Context) (bool, error)
	SetCanBuild(ctx context.Context, allow bool) error
	GetCanDig(ctx context.Context) (bool, error)
	SetCanDig(ctx context.Context, allow bool) error
	GetCanUseDoorsAndSwitches(ctx context.Context) (bool, error)
	SetCanUseDoorsAndSwitches(ctx context.Context, allow bool) error
	GetCanOpenContainers(ctx context.Context) (bool, error)
	SetCanOpenContainers(ctx context.Context, allow bool) error
	GetCanAttackPlayers(ctx context.Context) (bool, error)
	SetCanAttackPlayers(ctx context.Context, allow bool) error
	GetCanAttackMobs(ctx context.Context) (bool, error)
	SetCanAttackMobs(ctx context.Context, allow bool) error
	GetCanUseOperatorCommands(ctx context.Context) (bool, error)
	SetCanUseOperatorCommands(ctx context.Context, allow bool) error
	GetCanTeleport(ctx context.Context) (bool, error)
	SetCanTeleport(ctx context.Context, allow bool) (bool, error)

	GetStatusInvulnerable(ctx context.Context) (bool, error)
	GetStatusFlying(ctx context.Context) (bool, error)
	GetStatusMayFly(ctx context.Context) (bool, error)

	GetDeviceID(ctx context.Context) (string, error)
	GetEntityRuntimeID(ctx context.Context) (uint64, error)
	GetEntityMetadata(ctx context.Context) (map[uint32]any, error)

	GetIsOP(ctx context.Context) (bool, error)
	GetOnline(ctx context.Context) (bool, error)

	RawSay(jsonText string) error
	Say(message string) error
	Title(message string) error
	Subtitle(subtitleMessage, titleMessage string) error
	ActionBar(message string) error
}

