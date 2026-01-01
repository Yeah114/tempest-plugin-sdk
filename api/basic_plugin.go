package api

import (
	"context"

	"github.com/Yeah114/tempest-plugin-sdk/define"
)

type BasicPlugin struct {
	frame define.Frame
	config map[string]interface{}
}

func (p *BasicPlugin) Init(frame define.Frame, config map[string]interface{}) {
	p.frame = frame
	p.config = config
}

func (p *BasicPlugin) Frame() (frame define.Frame) {
	return p.frame
}

func (p *BasicPlugin) Config() (config map[string]interface{}) {
	return p.config
}

func (p *BasicPlugin) Load(ctx context.Context) (err error) {
	return
}

func (p *BasicPlugin) Unload(ctx context.Context) (err error) {
	return
}
