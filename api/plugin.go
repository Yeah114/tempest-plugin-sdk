package api

import (
	"context"

	"github.com/Yeah114/tempest/tempest-plugin-sdk/define"
)

// Plugin
type Plugin interface {
	Init(frame define.Frame, config map[string]interface{})
	Frame() (frame define.Frame)
	Config() (config map[string]interface{})
	Load(ctx context.Context) (err error)
	Unload(ctx context.Context) (err error)
}
