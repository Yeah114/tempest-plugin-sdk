package api

import (
	"context"

	"github.com/Yeah114/tempest-plugin-sdk/define"
)

// Plugin
type Plugin interface {
	Init(frame define.Frame, id string, config map[string]interface{})
	Frame() (frame define.Frame)
	ID() (id string)
	Config() (config map[string]interface{})
	Load(ctx context.Context) (err error)
	Unload(ctx context.Context) (err error)
}
