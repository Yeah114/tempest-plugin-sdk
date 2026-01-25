package api

import (
	"context"

	"github.com/Yeah114/EmptyDea-plugin-sdk/define"
)

const NameBrainModule = "brain"

type BrainModule interface {
	Name() string

	// 第1次启用的时候会调用 daemon 的 Load，之后重复启用仅会调用 ReConfig
	EnableDaemon(ctx context.Context, name string, config map[string]interface{}) (actualConfig map[string]interface{}, daemon define.Daemon, err error)
	DisableDaemon(ctx context.Context, name string) (err error)
}
