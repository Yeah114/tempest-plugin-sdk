package protocol

import (
	"context"
	"errors"
	"net"
	"net/rpc"
	"sync"

	"github.com/hashicorp/go-plugin"

	"github.com/Yeah114/tempest-plugin-sdk/api"
)

type PlayersModuleNameResp struct {
	Name string
}

type PlayersGetPlayerArgs struct {
	UUID     string
	Name     string
	TimeoutMs int64
}

type PlayersGetPlayerResp struct {
	Exists       bool
	PlayerBrokerID uint32
}

type PlayersGetAllOnlineArgs struct {
	TimeoutMs int64
}

type PlayersGetAllOnlineResp struct {
	PlayerBrokerIDs []uint32
}

type PlayersRegisterWhenPlayerChangeArgs struct {
	CallbackBrokerID uint32
}

type PlayersListenerResp struct {
	ListenerID string
}

type PlayersUnregisterArgs struct {
	ListenerID string
}

type PlayersChangeEventArgs struct {
	Event api.PlayerChangeEvent
}

type playersChangeCallbackServer struct {
	handler func(*api.PlayerChangeEvent)
}

func (s *playersChangeCallbackServer) OnEvent(args *PlayersChangeEventArgs, _ *Empty) error {
	if s == nil || s.handler == nil || args == nil {
		return nil
	}
	e := args.Event
	s.handler(&e)
	return nil
}

type playersChangeCallbackClient struct {
	c  *rpc.Client
	mu sync.Mutex
}

func (c *playersChangeCallbackClient) Close() error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Close()
}

func (c *playersChangeCallbackClient) OnEvent(event *api.PlayerChangeEvent) error {
	if c == nil || c.c == nil || event == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.OnEvent", &PlayersChangeEventArgs{Event: *event}, &Empty{})
}

type PlayersSayToArgs struct {
	Target   string
	Message  string
	JSONText string
	Subtitle string
	Title    string
}

type PlayersModuleRPCServer struct {
	Impl   api.PlayersModule
	broker *plugin.MuxBroker
}

func (s *PlayersModuleRPCServer) Name(_ *Empty, resp *PlayersModuleNameResp) error {
	if resp == nil {
		return nil
	}
	resp.Name = api.NamePlayersModule
	if s == nil || s.Impl == nil {
		return nil
	}
	resp.Name = s.Impl.Name()
	return nil
}

func (s *PlayersModuleRPCServer) NewPlayerKit(args *PlayersGetPlayerArgs, resp *PlayersGetPlayerResp) error {
	if resp == nil {
		return nil
	}
	resp.Exists = false
	resp.PlayerBrokerID = 0
	if s == nil || s.Impl == nil || s.broker == nil || args == nil {
		return nil
	}
	kit := s.Impl.NewPlayerKit(args.UUID)
	if kit == nil {
		return nil
	}
	id := s.broker.NextId()
	go acceptAndServeMuxBroker(s.broker, id, &PlayerKitRPCServer{Impl: kit})
	resp.Exists = true
	resp.PlayerBrokerID = id
	return nil
}

func (s *PlayersModuleRPCServer) GetPlayerByUUID(args *PlayersGetPlayerArgs, resp *PlayersGetPlayerResp) error {
	if resp == nil {
		return nil
	}
	resp.Exists = false
	resp.PlayerBrokerID = 0
	if s == nil || s.Impl == nil || s.broker == nil || args == nil {
		return nil
	}
	ctx, cancel := ctxFromTimeoutMs(args.TimeoutMs)
	defer cancel()
	kit, err := s.Impl.GetPlayerByUUID(ctx, args.UUID)
	if err != nil {
		return err
	}
	if kit == nil {
		return nil
	}
	id := s.broker.NextId()
	go acceptAndServeMuxBroker(s.broker, id, &PlayerKitRPCServer{Impl: kit})
	resp.Exists = true
	resp.PlayerBrokerID = id
	return nil
}

func (s *PlayersModuleRPCServer) GetPlayerByName(args *PlayersGetPlayerArgs, resp *PlayersGetPlayerResp) error {
	if resp == nil {
		return nil
	}
	resp.Exists = false
	resp.PlayerBrokerID = 0
	if s == nil || s.Impl == nil || s.broker == nil || args == nil {
		return nil
	}
	ctx, cancel := ctxFromTimeoutMs(args.TimeoutMs)
	defer cancel()
	kit, err := s.Impl.GetPlayerByName(ctx, args.Name)
	if err != nil {
		return err
	}
	if kit == nil {
		return nil
	}
	id := s.broker.NextId()
	go acceptAndServeMuxBroker(s.broker, id, &PlayerKitRPCServer{Impl: kit})
	resp.Exists = true
	resp.PlayerBrokerID = id
	return nil
}

func (s *PlayersModuleRPCServer) GetAllOnlinePlayers(args *PlayersGetAllOnlineArgs, resp *PlayersGetAllOnlineResp) error {
	if resp == nil {
		return nil
	}
	resp.PlayerBrokerIDs = nil
	if s == nil || s.Impl == nil || s.broker == nil {
		return nil
	}
	timeoutMs := int64(0)
	if args != nil {
		timeoutMs = args.TimeoutMs
	}
	ctx, cancel := ctxFromTimeoutMs(timeoutMs)
	defer cancel()

	kits, err := s.Impl.GetAllOnlinePlayers(ctx)
	if err != nil {
		return err
	}
	if len(kits) == 0 {
		resp.PlayerBrokerIDs = []uint32{}
		return nil
	}
	resp.PlayerBrokerIDs = make([]uint32, 0, len(kits))
	for _, kit := range kits {
		if kit == nil {
			continue
		}
		id := s.broker.NextId()
		go acceptAndServeMuxBroker(s.broker, id, &PlayerKitRPCServer{Impl: kit})
		resp.PlayerBrokerIDs = append(resp.PlayerBrokerIDs, id)
	}
	return nil
}

func (s *PlayersModuleRPCServer) RegisterWhenPlayerChange(args *PlayersRegisterWhenPlayerChangeArgs, resp *PlayersListenerResp) error {
	if resp == nil {
		return nil
	}
	resp.ListenerID = ""
	if s == nil || s.Impl == nil || s.broker == nil || args == nil {
		return nil
	}
	if args.CallbackBrokerID == 0 {
		return errors.New("PlayersModuleRPCServer.RegisterWhenPlayerChange: callback broker id is 0")
	}
	conn, err := s.broker.Dial(args.CallbackBrokerID)
	if err != nil {
		return err
	}
	cb := &playersChangeCallbackClient{c: rpc.NewClient(conn)}
	listenerID, err := s.Impl.RegisterWhenPlayerChange(func(event *api.PlayerChangeEvent) {
		_ = cb.OnEvent(event)
	})
	if err != nil {
		_ = cb.Close()
		return err
	}
	resp.ListenerID = listenerID
	return nil
}

func (s *PlayersModuleRPCServer) UnregisterWhenPlayerChange(args *PlayersUnregisterArgs, resp *BoolResp) error {
	if resp == nil {
		return nil
	}
	resp.OK = false
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	resp.OK = s.Impl.UnregisterWhenPlayerChange(args.ListenerID)
	return nil
}

func (s *PlayersModuleRPCServer) RawSayTo(args *PlayersSayToArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	return s.Impl.RawSayTo(args.Target, args.JSONText)
}

func (s *PlayersModuleRPCServer) SayTo(args *PlayersSayToArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	return s.Impl.SayTo(args.Target, args.Message)
}

func (s *PlayersModuleRPCServer) RawTitleTo(args *PlayersSayToArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	return s.Impl.RawTitleTo(args.Target, args.JSONText)
}

func (s *PlayersModuleRPCServer) TitleTo(args *PlayersSayToArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	return s.Impl.TitleTo(args.Target, args.Message)
}

func (s *PlayersModuleRPCServer) RawSubtitleTo(args *PlayersSayToArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	return s.Impl.RawSubtitleTo(args.Target, args.Subtitle, args.Title)
}

func (s *PlayersModuleRPCServer) SubtitleTo(args *PlayersSayToArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	return s.Impl.SubtitleTo(args.Target, args.Subtitle, args.Title)
}

func (s *PlayersModuleRPCServer) ActionBarTo(args *PlayersSayToArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	return s.Impl.ActionBarTo(args.Target, args.Message)
}

type playersModuleRPCClient struct {
	c      *rpc.Client
	broker *plugin.MuxBroker
	mu     sync.Mutex
}

func newPlayersModuleRPCClient(conn net.Conn, broker *plugin.MuxBroker) api.PlayersModule {
	if conn == nil {
		return nil
	}
	return &playersModuleRPCClient{c: rpc.NewClient(conn), broker: broker}
}

func (c *playersModuleRPCClient) Name() string { return api.NamePlayersModule }

func (c *playersModuleRPCClient) NewPlayerKit(uuid string) api.PlayerKit {
	if c == nil || c.c == nil || c.broker == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp PlayersGetPlayerResp
	_ = c.c.Call("Plugin.NewPlayerKit", &PlayersGetPlayerArgs{UUID: uuid}, &resp)
	if !resp.Exists || resp.PlayerBrokerID == 0 {
		return nil
	}
	conn, err := c.broker.Dial(resp.PlayerBrokerID)
	if err != nil {
		return nil
	}
	return newPlayerKitRPCClient(conn)
}

func (c *playersModuleRPCClient) GetAllOnlinePlayers(ctx context.Context) ([]api.PlayerKit, error) {
	if c == nil || c.c == nil || c.broker == nil {
		return nil, errors.New("playersModuleRPCClient.GetAllOnlinePlayers: client is not initialised")
	}
	timeoutMs := timeoutMsFromContext(ctx)
	c.mu.Lock()
	var resp PlayersGetAllOnlineResp
	err := c.c.Call("Plugin.GetAllOnlinePlayers", &PlayersGetAllOnlineArgs{TimeoutMs: timeoutMs}, &resp)
	c.mu.Unlock()
	if err != nil {
		return nil, err
	}
	out := make([]api.PlayerKit, 0, len(resp.PlayerBrokerIDs))
	for _, id := range resp.PlayerBrokerIDs {
		if id == 0 {
			continue
		}
		conn, dialErr := c.broker.Dial(id)
		if dialErr != nil {
			continue
		}
		out = append(out, newPlayerKitRPCClient(conn))
	}
	return out, nil
}

func (c *playersModuleRPCClient) GetPlayerByName(ctx context.Context, name string) (api.PlayerKit, error) {
	if c == nil || c.c == nil || c.broker == nil {
		return nil, errors.New("playersModuleRPCClient.GetPlayerByName: client is not initialised")
	}
	timeoutMs := timeoutMsFromContext(ctx)
	c.mu.Lock()
	var resp PlayersGetPlayerResp
	err := c.c.Call("Plugin.GetPlayerByName", &PlayersGetPlayerArgs{Name: name, TimeoutMs: timeoutMs}, &resp)
	c.mu.Unlock()
	if err != nil {
		return nil, err
	}
	if !resp.Exists || resp.PlayerBrokerID == 0 {
		return nil, nil
	}
	conn, err := c.broker.Dial(resp.PlayerBrokerID)
	if err != nil {
		return nil, err
	}
	return newPlayerKitRPCClient(conn), nil
}

func (c *playersModuleRPCClient) GetPlayerByUUID(ctx context.Context, uuid string) (api.PlayerKit, error) {
	if c == nil || c.c == nil || c.broker == nil {
		return nil, errors.New("playersModuleRPCClient.GetPlayerByUUID: client is not initialised")
	}
	timeoutMs := timeoutMsFromContext(ctx)
	c.mu.Lock()
	var resp PlayersGetPlayerResp
	err := c.c.Call("Plugin.GetPlayerByUUID", &PlayersGetPlayerArgs{UUID: uuid, TimeoutMs: timeoutMs}, &resp)
	c.mu.Unlock()
	if err != nil {
		return nil, err
	}
	if !resp.Exists || resp.PlayerBrokerID == 0 {
		return nil, nil
	}
	conn, err := c.broker.Dial(resp.PlayerBrokerID)
	if err != nil {
		return nil, err
	}
	return newPlayerKitRPCClient(conn), nil
}

func (c *playersModuleRPCClient) RegisterWhenPlayerChange(handler func(event *api.PlayerChangeEvent)) (string, error) {
	if c == nil || c.c == nil || c.broker == nil {
		return "", errors.New("playersModuleRPCClient.RegisterWhenPlayerChange: client is not initialised")
	}
	if handler == nil {
		return "", errors.New("playersModuleRPCClient.RegisterWhenPlayerChange: handler is nil")
	}
	cbID := c.broker.NextId()
	go acceptAndServeMuxBroker(c.broker, cbID, &playersChangeCallbackServer{handler: handler})

	c.mu.Lock()
	defer c.mu.Unlock()
	var resp PlayersListenerResp
	if err := c.c.Call("Plugin.RegisterWhenPlayerChange", &PlayersRegisterWhenPlayerChangeArgs{CallbackBrokerID: cbID}, &resp); err != nil {
		return "", err
	}
	return resp.ListenerID, nil
}

func (c *playersModuleRPCClient) UnregisterWhenPlayerChange(listenerID string) bool {
	if c == nil || c.c == nil {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp BoolResp
	if err := c.c.Call("Plugin.UnregisterWhenPlayerChange", &PlayersUnregisterArgs{ListenerID: listenerID}, &resp); err != nil {
		return false
	}
	return resp.OK
}

func (c *playersModuleRPCClient) RawSayTo(target string, jsonText string) error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.RawSayTo", &PlayersSayToArgs{Target: target, JSONText: jsonText}, &Empty{})
}

func (c *playersModuleRPCClient) SayTo(target string, message string) error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.SayTo", &PlayersSayToArgs{Target: target, Message: message}, &Empty{})
}

func (c *playersModuleRPCClient) RawTitleTo(target string, jsonText string) error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.RawTitleTo", &PlayersSayToArgs{Target: target, JSONText: jsonText}, &Empty{})
}

func (c *playersModuleRPCClient) TitleTo(target string, message string) error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.TitleTo", &PlayersSayToArgs{Target: target, Message: message}, &Empty{})
}

func (c *playersModuleRPCClient) RawSubtitleTo(target string, subtitleJsonText, titleJsonText string) error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.RawSubtitleTo", &PlayersSayToArgs{Target: target, Subtitle: subtitleJsonText, Title: titleJsonText}, &Empty{})
}

func (c *playersModuleRPCClient) SubtitleTo(target string, subtitleMessage, titleMessage string) error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.SubtitleTo", &PlayersSayToArgs{Target: target, Subtitle: subtitleMessage, Title: titleMessage}, &Empty{})
}

func (c *playersModuleRPCClient) ActionBarTo(target string, message string) error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.ActionBarTo", &PlayersSayToArgs{Target: target, Message: message}, &Empty{})
}

var _ api.PlayersModule = (*playersModuleRPCClient)(nil)

