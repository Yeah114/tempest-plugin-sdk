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
	Init(frame sdkdefine.Frame, id string, config map[string]interface{}) error
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
	ID            string
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
	Exists         bool
	Name           string
	ModuleKind     string
	ModuleBrokerID uint32
}

type GetPluginConfigArgs struct {
	ID string
}

type GetPluginConfigResp struct {
	Exists bool
	Config sdkdefine.PluginConfig
}

type RegisterWhenActivateArgs struct {
	CallbackBrokerID uint32
}

type RegisterWhenActivateResp struct {
	ListenerID string
}

type UnregisterWhenActivateArgs struct {
	ListenerID string
}

type UnregisterWhenActivateResp struct {
	OK bool
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
		go acceptAndServeMuxBroker(s.broker, id, &ChatModuleRPCServer{Impl: chatMod, broker: s.broker})
		resp.ModuleKind = api.NameChatModule
		resp.ModuleBrokerID = id
		return nil
	}
	if cmdsMod, ok := any(mod).(api.CommandsModule); ok {
		id := s.broker.NextId()
		go acceptAndServeMuxBroker(s.broker, id, &CommandsModuleRPCServer{Impl: cmdsMod})
		resp.ModuleKind = api.NameCommandsModule
		resp.ModuleBrokerID = id
		return nil
	}
	if flexMod, ok := any(mod).(api.FlexModule); ok {
		id := s.broker.NextId()
		go acceptAndServeMuxBroker(s.broker, id, &FlexModuleRPCServer{Impl: flexMod, broker: s.broker})
		resp.ModuleKind = api.NameFlexModule
		resp.ModuleBrokerID = id
		return nil
	}
	if uqMod, ok := any(mod).(api.UQHolderModule); ok {
		id := s.broker.NextId()
		go acceptAndServeMuxBroker(s.broker, id, &UQHolderModuleRPCServer{Impl: uqMod})
		resp.ModuleKind = api.NameUQHolderModule
		resp.ModuleBrokerID = id
		return nil
	}
	if gmMod, ok := any(mod).(api.GameMenuModule); ok {
		id := s.broker.NextId()
		go acceptAndServeMuxBroker(s.broker, id, &GameMenuModuleRPCServer{Impl: gmMod, broker: s.broker})
		resp.ModuleKind = api.NameGameMenuModule
		resp.ModuleBrokerID = id
		return nil
	}
	if tmMod, ok := any(mod).(api.TerminalMenuModule); ok {
		id := s.broker.NextId()
		go acceptAndServeMuxBroker(s.broker, id, &TerminalMenuModuleRPCServer{Impl: tmMod, broker: s.broker})
		resp.ModuleKind = api.NameTerminalMenuModule
		resp.ModuleBrokerID = id
		return nil
	}
	if terminalMod, ok := any(mod).(api.TerminalModule); ok {
		id := s.broker.NextId()
		go acceptAndServeMuxBroker(s.broker, id, &TerminalModuleRPCServer{Impl: terminalMod, broker: s.broker})
		resp.ModuleKind = api.NameTerminalModule
		resp.ModuleBrokerID = id
		return nil
	}
	if playersMod, ok := any(mod).(api.PlayersModule); ok {
		id := s.broker.NextId()
		go acceptAndServeMuxBroker(s.broker, id, &PlayersModuleRPCServer{Impl: playersMod, broker: s.broker})
		resp.ModuleKind = api.NamePlayersModule
		resp.ModuleBrokerID = id
		return nil
	}
	if loggerMod, ok := any(mod).(api.LoggerModule); ok {
		id := s.broker.NextId()
		go acceptAndServeMuxBroker(s.broker, id, &LoggerModuleRPCServer{Impl: loggerMod})
		resp.ModuleKind = api.NameLoggerModule
		resp.ModuleBrokerID = id
		return nil
	}
	if spMod, ok := any(mod).(api.StoragePathModule); ok {
		id := s.broker.NextId()
		go acceptAndServeMuxBroker(s.broker, id, &StoragePathModuleRPCServer{Impl: spMod})
		resp.ModuleKind = api.NameStoragePathModule
		resp.ModuleBrokerID = id
	}
	return nil
}

func (s *frameRPCServer) GetPluginConfig(args *GetPluginConfigArgs, resp *GetPluginConfigResp) error {
	if resp == nil {
		return nil
	}
	resp.Exists = false
	resp.Config = sdkdefine.PluginConfig{}
	if s == nil || s.Frame == nil || args == nil || args.ID == "" {
		return nil
	}
	cfg, ok := s.Frame.GetPluginConfig(args.ID)
	if !ok {
		return nil
	}
	resp.Exists = true
	resp.Config = cfg
	return nil
}

func (s *frameRPCServer) RegisterWhenActivate(args *RegisterWhenActivateArgs, resp *RegisterWhenActivateResp) error {
	if resp == nil {
		return nil
	}
	resp.ListenerID = ""
	if s == nil || s.Frame == nil || args == nil {
		return nil
	}
	if s.broker == nil || args.CallbackBrokerID == 0 {
		return nil
	}

	id, err := s.Frame.RegisterWhenActivate(func() {
		conn, dialErr := s.broker.Dial(args.CallbackBrokerID)
		if dialErr != nil || conn == nil {
			return
		}
		client := rpc.NewClient(conn)
		_ = client.Call("Plugin.Activate", &Empty{}, &Empty{})
		_ = client.Close()
	})
	if err != nil {
		return err
	}
	resp.ListenerID = id
	return nil
}

func (s *frameRPCServer) UnregisterWhenActivate(args *UnregisterWhenActivateArgs, resp *UnregisterWhenActivateResp) error {
	if resp == nil {
		return nil
	}
	resp.OK = false
	if s == nil || s.Frame == nil || args == nil || args.ListenerID == "" {
		return nil
	}
	resp.OK = s.Frame.UnregisterWhenActivate(args.ListenerID)
	return nil
}

type frameRPCClient struct {
	c      *rpc.Client
	broker *plugin.MuxBroker
	mu     sync.Mutex
}

type activateCallbackRPCServer struct {
	Handler func()
}

func (s *activateCallbackRPCServer) Activate(_ *Empty, _ *Empty) error {
	if s == nil || s.Handler == nil {
		return nil
	}
	s.Handler()
	return nil
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
			case api.NameCommandsModule:
				if m := newCommandsModuleRPCClient(conn); m != nil {
					return m, true
				}
			case api.NameFlexModule:
				if m := newFlexModuleRPCClient(conn, c.broker); m != nil {
					return m, true
				}
			case api.NameUQHolderModule:
				if m := newUQHolderModuleRPCClient(conn); m != nil {
					return m, true
				}
			case api.NameGameMenuModule:
				if m := newGameMenuModuleRPCClient(conn, c.broker); m != nil {
					return m, true
				}
			case api.NameTerminalMenuModule:
				if m := newTerminalMenuModuleRPCClient(conn, c.broker); m != nil {
					return m, true
				}
			case api.NameTerminalModule:
				if m := newTerminalModuleRPCClient(conn, c.broker); m != nil {
					return m, true
				}
			case api.NamePlayersModule:
				if m := newPlayersModuleRPCClient(conn, c.broker); m != nil {
					return m, true
				}
			case api.NameLoggerModule:
				if m := newLoggerModuleRPCClient(conn); m != nil {
					return m, true
				}
			case api.NameStoragePathModule:
				if m := newStoragePathModuleRPCClient(conn); m != nil {
					return m, true
				}
			default:
				_ = conn.Close()
			}
		}
	}

	return frameModuleStub{name: resp.Name}, true
}

func (c *frameRPCClient) GetPluginConfig(id string) (sdkdefine.PluginConfig, bool) {
	if c == nil || c.c == nil || id == "" {
		return sdkdefine.PluginConfig{}, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	var resp GetPluginConfigResp
	if err := c.c.Call("Plugin.GetPluginConfig", &GetPluginConfigArgs{ID: id}, &resp); err != nil {
		return sdkdefine.PluginConfig{}, false
	}
	if !resp.Exists {
		return sdkdefine.PluginConfig{}, false
	}
	return resp.Config, true
}

func (c *frameRPCClient) RegisterWhenActivate(handler func()) (string, error) {
	if c == nil || c.c == nil || handler == nil {
		return "", nil
	}
	if c.broker == nil {
		return "", nil
	}

	brokerID := c.broker.NextId()
	go acceptAndServeMuxBroker(c.broker, brokerID, &activateCallbackRPCServer{Handler: handler})

	c.mu.Lock()
	defer c.mu.Unlock()

	var resp RegisterWhenActivateResp
	if err := c.c.Call("Plugin.RegisterWhenActivate", &RegisterWhenActivateArgs{CallbackBrokerID: brokerID}, &resp); err != nil {
		return "", err
	}
	return resp.ListenerID, nil
}

func (c *frameRPCClient) UnregisterWhenActivate(listenerID string) bool {
	if c == nil || c.c == nil || listenerID == "" {
		return false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	var resp UnregisterWhenActivateResp
	if err := c.c.Call("Plugin.UnregisterWhenActivate", &UnregisterWhenActivateArgs{ListenerID: listenerID}, &resp); err != nil {
		return false
	}
	return resp.OK
}

func (s *rpcServer) Init(args *InitArgs, _ *Empty) error {
	if s == nil || s.Impl == nil {
		return nil
	}
	id := ""
	cfg := map[string]interface{}{}
	if args != nil {
		id = args.ID
		if args.Config != nil {
			cfg = args.Config
		}
	}
	var frame sdkdefine.Frame
	if args != nil && args.FrameBrokerID != 0 && s.broker != nil {
		if conn, err := s.broker.Dial(args.FrameBrokerID); err == nil && conn != nil {
			frame = &frameRPCClient{c: rpc.NewClient(conn), broker: s.broker}
		}
	}
	s.Impl.Init(frame, id, cfg)
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

func (c *rpcClient) Init(frame sdkdefine.Frame, id string, config map[string]interface{}) error {
	if c == nil || c.c == nil {
		return nil
	}
	var brokerID uint32
	if c.broker != nil {
		brokerID = c.broker.NextId()
		go acceptAndServeMuxBroker(c.broker, brokerID, &frameRPCServer{Frame: frame, broker: c.broker})
	}
	return c.c.Call("Plugin.Init", &InitArgs{ID: id, Config: config, FrameBrokerID: brokerID}, &Empty{})
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
