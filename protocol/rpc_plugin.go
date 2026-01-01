package protocol

import (
	"context"
	"net/rpc"

	"github.com/hashicorp/go-plugin"

	"github.com/Yeah114/tempest/tempest-plugin-sdk/api"
	sdkdefine "github.com/Yeah114/tempest/tempest-plugin-sdk/define"
)

// RPCPlugin is the minimal surface exposed over go-plugin for api.Plugin.
// It intentionally only supports Init/Load/Unload to keep the protocol small.
type RPCPlugin interface {
	Init(config map[string]interface{}) error
	Load() error
	Unload() error
}

// DynamicRPCPlugin is the go-plugin wrapper.
// - On the plugin (server) side, set Impl.
// - On the host (client) side, Impl is unused.
type DynamicRPCPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	Impl api.Plugin
}

type initArgs struct {
	Config map[string]interface{}
}

type empty struct{}

type rpcServer struct {
	Impl api.Plugin
}

type emptyFrame struct{}

func (emptyFrame) ListModules() map[string]sdkdefine.Module       { return nil }
func (emptyFrame) GetModule(name string) (sdkdefine.Module, bool) { return nil, false }

func (s *rpcServer) Init(args *initArgs, _ *empty) error {
	if s == nil || s.Impl == nil {
		return nil
	}
	cfg := map[string]interface{}{}
	if args != nil && args.Config != nil {
		cfg = args.Config
	}
	s.Impl.Init(emptyFrame{}, cfg)
	return nil
}

func (s *rpcServer) Load(_ *empty, _ *empty) error {
	if s == nil || s.Impl == nil {
		return nil
	}
	return s.Impl.Load(context.Background())
}

func (s *rpcServer) Unload(_ *empty, _ *empty) error {
	if s == nil || s.Impl == nil {
		return nil
	}
	return s.Impl.Unload(context.Background())
}

type rpcClient struct {
	c *rpc.Client
}

func (c *rpcClient) Init(config map[string]interface{}) error {
	return c.c.Call("Plugin.Init", &initArgs{Config: config}, &empty{})
}

func (c *rpcClient) Load() error {
	return c.c.Call("Plugin.Load", &empty{}, &empty{})
}

func (c *rpcClient) Unload() error {
	return c.c.Call("Plugin.Unload", &empty{}, &empty{})
}

func (p *DynamicRPCPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &rpcServer{Impl: p.Impl}, nil
}

func (p *DynamicRPCPlugin) Client(_ *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &rpcClient{c: c}, nil
}
