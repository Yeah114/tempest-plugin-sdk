package api

import (
	"context"

	"github.com/Yeah114/tempest/tempest-plugin-sdk/define"
)

type BasicPlugin struct {
	frame define.Frame
	config define.Config
}

func (p BasicPlugin) Init(frame define.Frame, config define.Config) {
	p.frame = frame
	p.config = config
}

func (p BasicPlugin) Frame() (frame define.Frame) {
	return p.frame
}

func (p BasicPlugin) Config() (config define.Config) {
	return p.config
}

func (p BasicPlugin) Metadata() (metadata define.Metadata) {
	return define.Metadata{
		Name:        "BasicPlugin",
		Description: "test",
		Author:      "tempest",
	}
}

func (p BasicPlugin) Load(ctx context.Context) (err error) {
	return
}

func (p BasicPlugin) Unload(ctx context.Context) (err error) {
	return
}
