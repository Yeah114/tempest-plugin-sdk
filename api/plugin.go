package api

import (
	"context"

	"github.com/Yeah114/tempest/tempest-plugin-sdk/define"
)

// Plugin
type Plugin interface {
	Init(frame define.Frame, config define.Config)
	Frame() (frame define.Frame)
	Config() (config define.Config)
	Metadata() (metadata define.Metadata)
	Load(ctx context.Context) (err error)
	Unload(ctx context.Context) (err error)
}
