package api

import "context"

const NameUQHolderModule = "uqholder"

type GameRule struct {
	CanBeModified bool
	Value         string
}

type UQHolderModule interface {
	Name() string

	BotName(ctx context.Context) (string, error)
	BotRuntimeID(ctx context.Context) (uint64, error)
	BotUniqueID(ctx context.Context) (int64, error)
	BotIdentity(ctx context.Context) (string, error)
	BotUUIDStr(ctx context.Context) (string, error)
	BotXUID(ctx context.Context) (string, error)

	BasicRaw(ctx context.Context) (map[string]any, error)

	CompressThreshold(ctx context.Context) (uint16, bool, error)
	WorldGameMode(ctx context.Context) (int32, bool, error)
	GameMode(ctx context.Context) (int32, bool, error)
	WorldDifficulty(ctx context.Context) (uint32, bool, error)
	Time(ctx context.Context) (int32, bool, error)
	DayTime(ctx context.Context) (int32, bool, error)
	DayTimePercent(ctx context.Context) (float32, bool, error)
	CurrentTick(ctx context.Context) (int64, bool, error)
	SyncRatio(ctx context.Context) (float32, bool, error)
	BotDimension(ctx context.Context) (int32, bool, error)
	BotPosition(ctx context.Context) ([3]float32, bool, error)
	BotPositionOutOfSyncTick(ctx context.Context) (int64, bool, error)
	ClientDimension(ctx context.Context) (int32, bool, error)
	ClientHotBarSlot(ctx context.Context) (byte, bool, error)
	ClientHoldingItem(ctx context.Context) (map[string]any, bool, error)
	GameRules(ctx context.Context) (map[string]GameRule, error)

	ExtendRaw(ctx context.Context) (map[string]any, error)
}
