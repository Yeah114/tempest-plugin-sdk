package protocol

import (
	"errors"
	"net"
	"net/rpc"
	"sync"

	"github.com/hashicorp/go-plugin"

	"github.com/Yeah114/tempest-plugin-sdk/api"
)

type ChunkDaemonNameResp struct {
	Name string
}

type ChunkRegisterWhenNewChunkArgs struct {
	CallbackBrokerID uint32
}

type ChunkListenerResp struct {
	ListenerID string
}

type ChunkUnregisterArgs struct {
	ListenerID string
}

type ChunkNewChunkEventArgs struct {
	Event api.ChunkNewChunkEvent
}

type chunkNewChunkCallbackServer struct {
	handler func(*api.ChunkNewChunkEvent)
}

func (s *chunkNewChunkCallbackServer) OnEvent(args *ChunkNewChunkEventArgs, _ *Empty) error {
	if s == nil || s.handler == nil || args == nil {
		return nil
	}
	e := args.Event
	s.handler(&e)
	return nil
}

type chunkNewChunkCallbackClient struct {
	c  *rpc.Client
	mu sync.Mutex
}

func (c *chunkNewChunkCallbackClient) Close() error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Close()
}

func (c *chunkNewChunkCallbackClient) OnEvent(event *api.ChunkNewChunkEvent) error {
	if c == nil || c.c == nil || event == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.OnEvent", &ChunkNewChunkEventArgs{Event: *event}, &Empty{})
}

type ChunkDaemonRPCServer struct {
	Impl   api.ChunkDaemon
	broker *plugin.MuxBroker
}

func (s *ChunkDaemonRPCServer) Name(_ *Empty, resp *ChunkDaemonNameResp) error {
	if resp == nil {
		return nil
	}
	resp.Name = ""
	if s == nil || s.Impl == nil {
		return nil
	}
	resp.Name = s.Impl.Name()
	return nil
}

func (s *ChunkDaemonRPCServer) ReConfig(args *DaemonReConfigArgs, _ *Empty) error {
	if s == nil || s.Impl == nil {
		return nil
	}
	cfg := map[string]interface{}{}
	if args != nil && args.Config != nil {
		cfg = args.Config
	}
	return s.Impl.ReConfig(cfg)
}

func (s *ChunkDaemonRPCServer) Config(_ *Empty, resp *DaemonConfigResp) error {
	if resp == nil {
		return nil
	}
	resp.Config = nil
	if s == nil || s.Impl == nil {
		return nil
	}
	cfg := s.Impl.Config()
	if cfg == nil {
		resp.Config = map[string]interface{}{}
		return nil
	}
	resp.Config = cfg
	return nil
}

func (s *ChunkDaemonRPCServer) RegisterWhenNewChunk(args *ChunkRegisterWhenNewChunkArgs, resp *ChunkListenerResp) error {
	if resp == nil {
		return nil
	}
	resp.ListenerID = ""
	if s == nil || s.Impl == nil || s.broker == nil || args == nil {
		return nil
	}
	if args.CallbackBrokerID == 0 {
		return errors.New("ChunkDaemonRPCServer.RegisterWhenNewChunk: callback broker id is 0")
	}
	conn, err := s.broker.Dial(args.CallbackBrokerID)
	if err != nil {
		return err
	}
	cb := &chunkNewChunkCallbackClient{c: rpc.NewClient(conn)}
	listenerID, err := s.Impl.RegisterWhenNewChunk(func(event *api.ChunkNewChunkEvent) {
		_ = cb.OnEvent(event)
	})
	if err != nil {
		_ = cb.Close()
		return err
	}
	resp.ListenerID = listenerID
	return nil
}

func (s *ChunkDaemonRPCServer) UnregisterWhenNewChunk(args *ChunkUnregisterArgs, resp *BoolResp) error {
	if resp == nil {
		return nil
	}
	resp.OK = false
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	resp.OK = s.Impl.UnregisterWhenNewChunk(args.ListenerID)
	return nil
}

type chunkDaemonRPCClient struct {
	c      *rpc.Client
	broker *plugin.MuxBroker
	mu     sync.Mutex
}

func newChunkDaemonRPCClient(conn net.Conn, broker *plugin.MuxBroker) api.ChunkDaemon {
	if conn == nil {
		return nil
	}
	return &chunkDaemonRPCClient{c: rpc.NewClient(conn), broker: broker}
}

func (c *chunkDaemonRPCClient) Name() string {
	if c == nil || c.c == nil {
		return ""
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp ChunkDaemonNameResp
	_ = c.c.Call("Plugin.Name", &Empty{}, &resp)
	return resp.Name
}

func (c *chunkDaemonRPCClient) ReConfig(config map[string]interface{}) error {
	if c == nil || c.c == nil {
		return errors.New("chunk daemon rpc client not initialised")
	}
	if config == nil {
		config = map[string]interface{}{}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.ReConfig", &DaemonReConfigArgs{Config: config}, &Empty{})
}

func (c *chunkDaemonRPCClient) Config() map[string]interface{} {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp DaemonConfigResp
	_ = c.c.Call("Plugin.Config", &Empty{}, &resp)
	if resp.Config == nil {
		return map[string]interface{}{}
	}
	return resp.Config
}

func (c *chunkDaemonRPCClient) RegisterWhenNewChunk(handler func(event *api.ChunkNewChunkEvent)) (string, error) {
	if c == nil || c.c == nil {
		return "", errors.New("chunk daemon rpc client not initialised")
	}
	if c.broker == nil {
		return "", errors.New("chunk daemon rpc client broker is nil")
	}
	if handler == nil {
		return "", errors.New("handler is nil")
	}

	bid := c.broker.NextId()
	go acceptAndServeMuxBroker(c.broker, bid, &chunkNewChunkCallbackServer{handler: handler})

	c.mu.Lock()
	defer c.mu.Unlock()
	var resp ChunkListenerResp
	if err := c.c.Call("Plugin.RegisterWhenNewChunk", &ChunkRegisterWhenNewChunkArgs{CallbackBrokerID: bid}, &resp); err != nil {
		return "", err
	}
	return resp.ListenerID, nil
}

func (c *chunkDaemonRPCClient) UnregisterWhenNewChunk(listenerID string) bool {
	if c == nil || c.c == nil {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp BoolResp
	_ = c.c.Call("Plugin.UnregisterWhenNewChunk", &ChunkUnregisterArgs{ListenerID: listenerID}, &resp)
	return resp.OK
}
