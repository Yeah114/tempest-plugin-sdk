package api

import (
	"context"

	"github.com/Yeah114/tempest-plugin-sdk/define"
)

type BrainModule interface {
	Name() string

	EnableDaemon(ctx context.Context, name string, config map[string]interface{}) (actualConfig map[string]interface{}, daemon define.Daemon, err error)
	DisableDaemon(ctx context.Context, name string) (err error)
}
