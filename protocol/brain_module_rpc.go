package protocol

import (
	"context"
	"errors"
	"net"
	"net/rpc"
	"sync"

	"github.com/hashicorp/go-plugin"

	"github.com/Yeah114/tempest-plugin-sdk/api"
	sdkdefine "github.com/Yeah114/tempest-plugin-sdk/define"
)

type BrainModuleNameResp struct {
	Name string
}

type BrainEnableDaemonArgs struct {
	Name      string
	Config    map[string]interface{}
	TimeoutMs int64
}

type BrainEnableDaemonResp struct {
	ActualConfig   map[string]interface{}
	DaemonExists   bool
	DaemonKind     string
	DaemonBrokerID uint32
}

type BrainDisableDaemonArgs struct {
	Name      string
	TimeoutMs int64
}

type BrainModuleRPCServer struct {
	Impl   api.BrainModule
	broker *plugin.MuxBroker
}

func (s *BrainModuleRPCServer) Name(_ *Empty, resp *BrainModuleNameResp) error {
	if resp == nil {
		return nil
	}
	resp.Name = api.NameBrainModule
	if s == nil || s.Impl == nil {
		return nil
	}
	resp.Name = s.Impl.Name()
	return nil
}

func (s *BrainModuleRPCServer) EnableDaemon(args *BrainEnableDaemonArgs, resp *BrainEnableDaemonResp) error {
	if resp == nil {
		return nil
	}
	resp.ActualConfig = nil
	resp.DaemonExists = false
	resp.DaemonKind = ""
	resp.DaemonBrokerID = 0

	if s == nil || s.Impl == nil || s.broker == nil || args == nil {
		return nil
	}

	ctx, cancel := ctxFromTimeoutMs(args.TimeoutMs)
	defer cancel()

	cfg := map[string]interface{}{}
	if args.Config != nil {
		cfg = args.Config
	}
	actual, dmn, err := s.Impl.EnableDaemon(ctx, args.Name, cfg)
	if err != nil {
		return err
	}
	if actual == nil {
		actual = map[string]interface{}{}
	}
	resp.ActualConfig = actual
	if dmn == nil {
		return nil
	}

	bid := s.broker.NextId()

	// Prefer richer daemon interfaces when available.
	if sb, ok := any(dmn).(api.ScoreboardDaemon); ok {
		resp.DaemonExists = true
		resp.DaemonKind = "scoreboard"
		resp.DaemonBrokerID = bid
		go acceptAndServeMuxBroker(s.broker, bid, &ScoreboardDaemonRPCServer{Impl: sb, broker: s.broker})
		return nil
	}
	if ck, ok := any(dmn).(api.ChunkDaemon); ok {
		resp.DaemonExists = true
		resp.DaemonKind = "chunk"
		resp.DaemonBrokerID = bid
		go acceptAndServeMuxBroker(s.broker, bid, &ChunkDaemonRPCServer{Impl: ck, broker: s.broker})
		return nil
	}

	resp.DaemonExists = true
	resp.DaemonKind = dmn.Name()
	resp.DaemonBrokerID = bid
	go acceptAndServeMuxBroker(s.broker, bid, &DaemonRPCServer{Impl: dmn})
	return nil
}

func (s *BrainModuleRPCServer) DisableDaemon(args *BrainDisableDaemonArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	ctx, cancel := ctxFromTimeoutMs(args.TimeoutMs)
	defer cancel()
	return s.Impl.DisableDaemon(ctx, args.Name)
}

type brainModuleRPCClient struct {
	c      *rpc.Client
	broker *plugin.MuxBroker
	mu     sync.Mutex
}

func newBrainModuleRPCClient(conn net.Conn, broker *plugin.MuxBroker) api.BrainModule {
	if conn == nil {
		return nil
	}
	return &brainModuleRPCClient{c: rpc.NewClient(conn), broker: broker}
}

func (c *brainModuleRPCClient) Name() string {
	if c == nil || c.c == nil {
		return api.NameBrainModule
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp BrainModuleNameResp
	_ = c.c.Call("Plugin.Name", &Empty{}, &resp)
	if resp.Name == "" {
		return api.NameBrainModule
	}
	return resp.Name
}

func (c *brainModuleRPCClient) EnableDaemon(ctx context.Context, name string, config map[string]interface{}) (map[string]interface{}, sdkdefine.Daemon, error) {
	if c == nil || c.c == nil {
		return nil, nil, errors.New("brainModuleRPCClient.EnableDaemon: client is not initialised")
	}
	if c.broker == nil {
		return nil, nil, errors.New("brainModuleRPCClient.EnableDaemon: broker is nil")
	}
	timeoutMs := int64(0)
	if ctx != nil {
		timeoutMs = timeoutMsFromCtx(ctx)
	}
	if config == nil {
		config = map[string]interface{}{}
	}

	c.mu.Lock()
	var resp BrainEnableDaemonResp
	err := c.c.Call("Plugin.EnableDaemon", &BrainEnableDaemonArgs{Name: name, Config: config, TimeoutMs: timeoutMs}, &resp)
	c.mu.Unlock()
	if err != nil {
		return nil, nil, err
	}
	if resp.ActualConfig == nil {
		resp.ActualConfig = map[string]interface{}{}
	}
	if !resp.DaemonExists || resp.DaemonBrokerID == 0 {
		return resp.ActualConfig, nil, nil
	}

	conn, err := c.broker.Dial(resp.DaemonBrokerID)
	if err != nil {
		return resp.ActualConfig, nil, err
	}

	if isDaemonKindScoreboard(resp.DaemonKind) {
		if d := newScoreboardDaemonRPCClient(conn, c.broker); d != nil {
			return resp.ActualConfig, d, nil
		}
		_ = conn.Close()
		return resp.ActualConfig, nil, errors.New("brainModuleRPCClient.EnableDaemon: failed to create scoreboard daemon client")
	}
	if isDaemonKindChunk(resp.DaemonKind) {
		if d := newChunkDaemonRPCClient(conn, c.broker); d != nil {
			return resp.ActualConfig, d, nil
		}
		_ = conn.Close()
		return resp.ActualConfig, nil, errors.New("brainModuleRPCClient.EnableDaemon: failed to create chunk daemon client")
	}

	if d := newDaemonRPCClient(conn); d != nil {
		return resp.ActualConfig, d, nil
	}
	_ = conn.Close()
	return resp.ActualConfig, nil, errors.New("brainModuleRPCClient.EnableDaemon: failed to create daemon client")
}

func (c *brainModuleRPCClient) DisableDaemon(ctx context.Context, name string) error {
	if c == nil || c.c == nil {
		return errors.New("brainModuleRPCClient.DisableDaemon: client is not initialised")
	}
	timeoutMs := int64(0)
	if ctx != nil {
		timeoutMs = timeoutMsFromCtx(ctx)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.DisableDaemon", &BrainDisableDaemonArgs{Name: name, TimeoutMs: timeoutMs}, &Empty{})
}
