package api

import (
	"context"

	"github.com/Yeah114/tempest-plugin-sdk/define"
)

type BasicPlugin struct {
	frame  define.Frame
	id     string
	config map[string]interface{}
}

func (p *BasicPlugin) Init(frame define.Frame, id string, config map[string]interface{}) {
	p.frame = frame
	p.id = id
	p.config = config
}

func (p *BasicPlugin) Frame() (frame define.Frame) {
	return p.frame
}

func (p *BasicPlugin) ID() (id string) {
	return p.id
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
