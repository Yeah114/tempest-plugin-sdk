package protocol

import (
	"errors"
	"net"
	"net/rpc"
	"strings"
	"sync"

	sdkdefine "github.com/Yeah114/tempest-plugin-sdk/define"
)

type DaemonNameResp struct {
	Name string
}

type DaemonReConfigArgs struct {
	Config map[string]interface{}
}

type DaemonConfigResp struct {
	Config map[string]interface{}
}

type DaemonRPCServer struct {
	Impl sdkdefine.Daemon
}

func (s *DaemonRPCServer) Name(_ *Empty, resp *DaemonNameResp) error {
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

func (s *DaemonRPCServer) ReConfig(args *DaemonReConfigArgs, _ *Empty) error {
	if s == nil || s.Impl == nil {
		return nil
	}
	cfg := map[string]interface{}{}
	if args != nil && args.Config != nil {
		cfg = args.Config
	}
	return s.Impl.ReConfig(cfg)
}

func (s *DaemonRPCServer) Config(_ *Empty, resp *DaemonConfigResp) error {
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

type daemonRPCClient struct {
	c  *rpc.Client
	mu sync.Mutex
}

func newDaemonRPCClient(conn net.Conn) sdkdefine.Daemon {
	if conn == nil {
		return nil
	}
	return &daemonRPCClient{c: rpc.NewClient(conn)}
}

func (c *daemonRPCClient) Name() string {
	if c == nil || c.c == nil {
		return ""
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp DaemonNameResp
	_ = c.c.Call("Plugin.Name", &Empty{}, &resp)
	return resp.Name
}

func (c *daemonRPCClient) ReConfig(config map[string]interface{}) error {
	if c == nil || c.c == nil {
		return errors.New("daemon rpc client not initialised")
	}
	if config == nil {
		config = map[string]interface{}{}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.ReConfig", &DaemonReConfigArgs{Config: config}, &Empty{})
}

func (c *daemonRPCClient) Config() map[string]interface{} {
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

func isDaemonKindScoreboard(kind string) bool {
	return strings.EqualFold(strings.TrimSpace(kind), "scoreboard")
}
