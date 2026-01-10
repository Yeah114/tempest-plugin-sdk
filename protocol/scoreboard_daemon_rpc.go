package protocol

import (
	"errors"
	"net"
	"net/rpc"
	"sync"

	"github.com/hashicorp/go-plugin"

	"github.com/Yeah114/tempest-plugin-sdk/api"
)

type ScoreboardDaemonNameResp struct {
	Name string
}

type ScoreboardRegisterWhenUpdateArgs struct {
	CallbackBrokerID uint32
}

type ScoreboardListenerResp struct {
	ListenerID string
}

type ScoreboardUnregisterArgs struct {
	ListenerID string
}

type ScoreboardUpdateEventArgs struct {
	Event api.ScoreUpdateEvent
}

type scoreboardUpdateCallbackServer struct {
	handler func(*api.ScoreUpdateEvent)
}

func (s *scoreboardUpdateCallbackServer) OnEvent(args *ScoreboardUpdateEventArgs, _ *Empty) error {
	if s == nil || s.handler == nil || args == nil {
		return nil
	}
	e := args.Event
	s.handler(&e)
	return nil
}

type scoreboardUpdateCallbackClient struct {
	c  *rpc.Client
	mu sync.Mutex
}

func (c *scoreboardUpdateCallbackClient) Close() error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Close()
}

func (c *scoreboardUpdateCallbackClient) OnEvent(event *api.ScoreUpdateEvent) error {
	if c == nil || c.c == nil || event == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.OnEvent", &ScoreboardUpdateEventArgs{Event: *event}, &Empty{})
}

type ScoreboardQueryScoreArgs struct {
	UUID string
}

type ScoreboardQueryScoreResp struct {
	Results []api.PlayerScoreQueryResult
}

type ScoreboardQueryRankArgs struct {
	ScoreboardName string
	Descending     bool
	MaxCount       int
}

type ScoreboardQueryRankResp struct {
	Results []api.RankQueryResult
}

type ScoreboardDaemonRPCServer struct {
	Impl   api.ScoreboardDaemon
	broker *plugin.MuxBroker
}

func (s *ScoreboardDaemonRPCServer) Name(_ *Empty, resp *ScoreboardDaemonNameResp) error {
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

func (s *ScoreboardDaemonRPCServer) ReConfig(args *DaemonReConfigArgs, _ *Empty) error {
	if s == nil || s.Impl == nil {
		return nil
	}
	cfg := map[string]interface{}{}
	if args != nil && args.Config != nil {
		cfg = args.Config
	}
	return s.Impl.ReConfig(cfg)
}

func (s *ScoreboardDaemonRPCServer) Config(_ *Empty, resp *DaemonConfigResp) error {
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

func (s *ScoreboardDaemonRPCServer) RegisterWhenScoreUpdate(args *ScoreboardRegisterWhenUpdateArgs, resp *ScoreboardListenerResp) error {
	if resp == nil {
		return nil
	}
	resp.ListenerID = ""
	if s == nil || s.Impl == nil || s.broker == nil || args == nil {
		return nil
	}
	if args.CallbackBrokerID == 0 {
		return errors.New("ScoreboardDaemonRPCServer.RegisterWhenScoreUpdate: callback broker id is 0")
	}
	conn, err := s.broker.Dial(args.CallbackBrokerID)
	if err != nil {
		return err
	}
	cb := &scoreboardUpdateCallbackClient{c: rpc.NewClient(conn)}
	listenerID, err := s.Impl.RegisterWhenScoreUpdate(func(event *api.ScoreUpdateEvent) {
		_ = cb.OnEvent(event)
	})
	if err != nil {
		_ = cb.Close()
		return err
	}
	resp.ListenerID = listenerID
	return nil
}

func (s *ScoreboardDaemonRPCServer) UnregisterWhenScoreUpdate(args *ScoreboardUnregisterArgs, resp *BoolResp) error {
	if resp == nil {
		return nil
	}
	resp.OK = false
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	resp.OK = s.Impl.UnregisterWhenScoreUpdate(args.ListenerID)
	return nil
}

func (s *ScoreboardDaemonRPCServer) QueryScoreByPlayerUUID(args *ScoreboardQueryScoreArgs, resp *ScoreboardQueryScoreResp) error {
	if resp == nil {
		return nil
	}
	resp.Results = nil
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	results := s.Impl.QueryScoreByPlayerUUID(args.UUID)
	if results == nil {
		resp.Results = []api.PlayerScoreQueryResult{}
		return nil
	}
	resp.Results = append([]api.PlayerScoreQueryResult(nil), (*results)...)
	return nil
}

func (s *ScoreboardDaemonRPCServer) QueryRankByScoreboard(args *ScoreboardQueryRankArgs, resp *ScoreboardQueryRankResp) error {
	if resp == nil {
		return nil
	}
	resp.Results = nil
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	results := s.Impl.QueryRankByScoreboard(args.ScoreboardName, args.Descending, args.MaxCount)
	if results == nil {
		resp.Results = []api.RankQueryResult{}
		return nil
	}
	resp.Results = append([]api.RankQueryResult(nil), (*results)...)
	return nil
}

type scoreboardDaemonRPCClient struct {
	c      *rpc.Client
	broker *plugin.MuxBroker
	mu     sync.Mutex
}

func newScoreboardDaemonRPCClient(conn net.Conn, broker *plugin.MuxBroker) api.ScoreboardDaemon {
	if conn == nil {
		return nil
	}
	return &scoreboardDaemonRPCClient{c: rpc.NewClient(conn), broker: broker}
}

func (c *scoreboardDaemonRPCClient) Name() string {
	if c == nil || c.c == nil {
		return ""
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp ScoreboardDaemonNameResp
	_ = c.c.Call("Plugin.Name", &Empty{}, &resp)
	return resp.Name
}

func (c *scoreboardDaemonRPCClient) ReConfig(config map[string]interface{}) error {
	if c == nil || c.c == nil {
		return errors.New("scoreboard daemon rpc client not initialised")
	}
	if config == nil {
		config = map[string]interface{}{}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.ReConfig", &DaemonReConfigArgs{Config: config}, &Empty{})
}

func (c *scoreboardDaemonRPCClient) Config() map[string]interface{} {
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

func (c *scoreboardDaemonRPCClient) RegisterWhenScoreUpdate(handler func(event *api.ScoreUpdateEvent)) (string, error) {
	if c == nil || c.c == nil {
		return "", errors.New("scoreboard daemon rpc client not initialised")
	}
	if handler == nil {
		return "", errors.New("scoreboardDaemonRPCClient.RegisterWhenScoreUpdate: handler is nil")
	}
	if c.broker == nil {
		return "", errors.New("scoreboardDaemonRPCClient.RegisterWhenScoreUpdate: broker is nil")
	}

	cbID := c.broker.NextId()
	go acceptAndServeMuxBroker(c.broker, cbID, &scoreboardUpdateCallbackServer{handler: handler})

	c.mu.Lock()
	defer c.mu.Unlock()
	var resp ScoreboardListenerResp
	if err := c.c.Call("Plugin.RegisterWhenScoreUpdate", &ScoreboardRegisterWhenUpdateArgs{CallbackBrokerID: cbID}, &resp); err != nil {
		return "", err
	}
	return resp.ListenerID, nil
}

func (c *scoreboardDaemonRPCClient) UnregisterWhenScoreUpdate(listenerID string) bool {
	if c == nil || c.c == nil || listenerID == "" {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp BoolResp
	if err := c.c.Call("Plugin.UnregisterWhenScoreUpdate", &ScoreboardUnregisterArgs{ListenerID: listenerID}, &resp); err != nil {
		return false
	}
	return resp.OK
}

func (c *scoreboardDaemonRPCClient) QueryScoreByPlayerUUID(uuid string) *[]api.PlayerScoreQueryResult {
	if c == nil || c.c == nil {
		out := []api.PlayerScoreQueryResult{}
		return &out
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp ScoreboardQueryScoreResp
	_ = c.c.Call("Plugin.QueryScoreByPlayerUUID", &ScoreboardQueryScoreArgs{UUID: uuid}, &resp)
	if resp.Results == nil {
		resp.Results = []api.PlayerScoreQueryResult{}
	}
	return &resp.Results
}

func (c *scoreboardDaemonRPCClient) QueryRankByScoreboard(scoreboardName string, descending bool, maxCount int) *[]api.RankQueryResult {
	if c == nil || c.c == nil {
		out := []api.RankQueryResult{}
		return &out
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp ScoreboardQueryRankResp
	_ = c.c.Call("Plugin.QueryRankByScoreboard", &ScoreboardQueryRankArgs{ScoreboardName: scoreboardName, Descending: descending, MaxCount: maxCount}, &resp)
	if resp.Results == nil {
		resp.Results = []api.RankQueryResult{}
	}
	return &resp.Results
}
