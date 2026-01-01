package protocol

import (
	"context"
	"net/rpc"
	"sync"

	"github.com/hashicorp/go-plugin"

	"github.com/Yeah114/tempest-plugin-sdk/api"
	sdkdefine "github.com/Yeah114/tempest-plugin-sdk/define"
)

// RPCPlugin is the minimal surface exposed over go-plugin for api.Plugin.
// It intentionally only supports Init/Load/Unload to keep the protocol small.
type RPCPlugin interface {
	Init(frame sdkdefine.Frame, config map[string]interface{}) error
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

type InitArgs struct {
	Config        map[string]interface{}
	FrameBrokerID uint32
}

type Empty struct{}

type rpcServer struct {
	Impl   api.Plugin
	broker *plugin.MuxBroker
}

type frameModuleStub struct {
	name string
}

func (m frameModuleStub) Name() string { return m.name }

type ListModulesResp struct {
	Names []string
}

type GetModuleArgs struct {
	Name string
}

type GetModuleResp struct {
	Exists        bool
	Name          string
	ModuleKind    string
	ModuleBrokerID uint32
}

type frameRPCServer struct {
	Frame  sdkdefine.Frame
	broker *plugin.MuxBroker
}

func (s *frameRPCServer) ListModules(_ *Empty, resp *ListModulesResp) error {
	if resp == nil {
		return nil
	}
	resp.Names = nil
	if s == nil || s.Frame == nil {
		return nil
	}
	mods := s.Frame.ListModules()
	if mods == nil {
		return nil
	}
	resp.Names = make([]string, 0, len(mods))
	for name := range mods {
		resp.Names = append(resp.Names, name)
	}
	return nil
}

func (s *frameRPCServer) GetModule(args *GetModuleArgs, resp *GetModuleResp) error {
	if resp == nil {
		return nil
	}
	resp.Exists = false
	resp.Name = ""
	resp.ModuleKind = ""
	resp.ModuleBrokerID = 0
	if s == nil || s.Frame == nil || args == nil {
		return nil
	}
	mod, ok := s.Frame.GetModule(args.Name)
	if !ok || mod == nil {
		return nil
	}
	resp.Exists = true
	resp.Name = mod.Name()

	if s.broker == nil {
		return nil
	}
	if chatMod, ok := any(mod).(api.ChatModule); ok {
		id := s.broker.NextId()
		go s.broker.AcceptAndServe(id, &ChatModuleRPCServer{Impl: chatMod, broker: s.broker})
		resp.ModuleKind = api.NameChatModule
		resp.ModuleBrokerID = id
	}
	return nil
}

type frameRPCClient struct {
	c      *rpc.Client
	broker *plugin.MuxBroker
	mu     sync.Mutex
}

func (c *frameRPCClient) ListModules() map[string]sdkdefine.Module {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	var resp ListModulesResp
	if err := c.c.Call("Plugin.ListModules", &Empty{}, &resp); err != nil {
		return nil
	}
	out := make(map[string]sdkdefine.Module, len(resp.Names))
	for _, name := range resp.Names {
		if name == "" {
			continue
		}
		out[name] = frameModuleStub{name: name}
	}
	return out
}

func (c *frameRPCClient) GetModule(name string) (sdkdefine.Module, bool) {
	if c == nil || c.c == nil {
		return nil, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	var resp GetModuleResp
	if err := c.c.Call("Plugin.GetModule", &GetModuleArgs{Name: name}, &resp); err != nil {
		return nil, false
	}
	if !resp.Exists || resp.Name == "" {
		return nil, false
	}

	if resp.ModuleBrokerID != 0 && c.broker != nil {
		if conn, err := c.broker.Dial(resp.ModuleBrokerID); err == nil && conn != nil {
			switch resp.ModuleKind {
			case api.NameChatModule:
				if m := newChatModuleRPCClient(conn, c.broker); m != nil {
					return m, true
				}
			default:
				_ = conn.Close()
			}
		}
	}

	return frameModuleStub{name: resp.Name}, true
}

func (s *rpcServer) Init(args *InitArgs, _ *Empty) error {
	if s == nil || s.Impl == nil {
		return nil
	}
	cfg := map[string]interface{}{}
	if args != nil && args.Config != nil {
		cfg = args.Config
	}
	var frame sdkdefine.Frame
	if args != nil && args.FrameBrokerID != 0 && s.broker != nil {
		if conn, err := s.broker.Dial(args.FrameBrokerID); err == nil && conn != nil {
			frame = &frameRPCClient{c: rpc.NewClient(conn), broker: s.broker}
		}
	}
	s.Impl.Init(frame, cfg)
	return nil
}

func (s *rpcServer) Load(_ *Empty, _ *Empty) error {
	if s == nil || s.Impl == nil {
		return nil
	}
	return s.Impl.Load(context.Background())
}

func (s *rpcServer) Unload(_ *Empty, _ *Empty) error {
	if s == nil || s.Impl == nil {
		return nil
	}
	return s.Impl.Unload(context.Background())
}

type rpcClient struct {
	c      *rpc.Client
	broker *plugin.MuxBroker
}

func (c *rpcClient) Init(frame sdkdefine.Frame, config map[string]interface{}) error {
	if c == nil || c.c == nil {
		return nil
	}
	var brokerID uint32
	if c.broker != nil {
		brokerID = c.broker.NextId()
		go c.broker.AcceptAndServe(brokerID, &frameRPCServer{Frame: frame, broker: c.broker})
	}
	return c.c.Call("Plugin.Init", &InitArgs{Config: config, FrameBrokerID: brokerID}, &Empty{})
}

func (c *rpcClient) Load() error {
	return c.c.Call("Plugin.Load", &Empty{}, &Empty{})
}

func (c *rpcClient) Unload() error {
	return c.c.Call("Plugin.Unload", &Empty{}, &Empty{})
}

func (p *DynamicRPCPlugin) Server(b *plugin.MuxBroker) (interface{}, error) {
	return &rpcServer{Impl: p.Impl, broker: b}, nil
}

func (p *DynamicRPCPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &rpcClient{c: c, broker: b}, nil
}
