package protocol

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"sync"
	"sync/atomic"

	"github.com/hashicorp/go-plugin"

	"github.com/Yeah114/EmptyDea-plugin-sdk/api"
)

type GameMenuModuleNameResp struct {
	Name string
}

// GameMenuEntryWire is the RPC-safe representation of api.GameMenuEntry.
// NOTE: api.GameMenuEntry contains a func field (OnTrigger) which is not gob-encodable.
type GameMenuEntryWire struct {
	Triggers     []string
	ArgumentHint string
	Usage        string
}

func toGameMenuEntryWire(entry *api.GameMenuEntry) GameMenuEntryWire {
	if entry == nil {
		return GameMenuEntryWire{}
	}
	return GameMenuEntryWire{
		Triggers:     append([]string(nil), entry.Triggers...),
		ArgumentHint: entry.ArgumentHint,
		Usage:        entry.Usage,
	}
}

func fromGameMenuEntryWire(w GameMenuEntryWire) api.GameMenuEntry {
	return api.GameMenuEntry{
		Triggers:     append([]string(nil), w.Triggers...),
		ArgumentHint: w.ArgumentHint,
		Usage:        w.Usage,
	}
}

type GameMenuEntryEvent struct {
	Info api.GameMenuEntryInfo
}

type GameMenuEntryCallbackServer struct {
	mu     sync.Mutex
	closed bool
	ch     chan<- *api.GameMenuEntryInfo
}

func (s *GameMenuEntryCallbackServer) OnEntry(args *GameMenuEntryEvent, _ *Empty) error {
	if s == nil {
		return nil
	}
	var info *api.GameMenuEntryInfo
	if args != nil {
		clean := args.Info
		clean.Triggers = append([]string(nil), clean.Triggers...)
		info = &clean
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.ch == nil || info == nil {
		return nil
	}
	select {
	case s.ch <- info:
	default:
	}
	return nil
}

func (s *GameMenuEntryCallbackServer) Stop(_ *Empty, _ *Empty) error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	if s.ch != nil {
		close(s.ch)
		s.ch = nil
	}
	return nil
}

type GameMenuEntryCallbackClient struct {
	c  *rpc.Client
	mu sync.Mutex
}

func (c *GameMenuEntryCallbackClient) Close() error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Close()
}

func (c *GameMenuEntryCallbackClient) OnEntry(info *api.GameMenuEntryInfo) error {
	if c == nil || c.c == nil || info == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	clean := *info
	clean.Triggers = append([]string(nil), info.Triggers...)
	return c.c.Call("Plugin.OnEntry", &GameMenuEntryEvent{Info: clean}, &Empty{})
}

func (c *GameMenuEntryCallbackClient) Stop() {
	if c == nil || c.c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.c.Call("Plugin.Stop", &Empty{}, &Empty{})
}

type GameMenuTriggerEvent struct {
	Chat *api.ChatMsg
}

type GameMenuTriggerCallbackServer struct {
	handler func(chat *api.ChatMsg)
}

func (s *GameMenuTriggerCallbackServer) OnTrigger(args *GameMenuTriggerEvent, _ *Empty) error {
	if s == nil || s.handler == nil {
		return nil
	}
	s.handler(sanitizeChatMsgForRPC(nilSafeChatMsg(args)))
	return nil
}

type GameMenuTriggerCallbackClient struct {
	c  *rpc.Client
	mu sync.Mutex
}

func (c *GameMenuTriggerCallbackClient) Close() error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Close()
}

func (c *GameMenuTriggerCallbackClient) OnTrigger(chat *api.ChatMsg) error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.OnTrigger", &GameMenuTriggerEvent{Chat: sanitizeChatMsgForRPC(chat)}, &Empty{})
}

type GameMenuRegisterEntryArgs struct {
	Entry           GameMenuEntryWire
	TriggerBrokerID uint32
}

type GameMenuRegisterEntryResp struct {
	EntryID string
}

type GameMenuRemoveArgs struct {
	EntryID string
}

type GameMenuSubscribeArgs struct {
	CallbackBrokerID uint32
}

type GameMenuSubscribeResp struct {
	SubID string
}

type GameMenuUnsubscribeArgs struct {
	SubID string
}

type GameMenuTriggerArgs struct {
	EntryID string
	Chat    *api.ChatMsg
}

type GameMenuModuleRPCServer struct {
	Impl   api.GameMenuModule
	broker *plugin.MuxBroker

	mu           sync.Mutex
	subs         map[string]context.CancelFunc
	subCallbacks map[string]*GameMenuEntryCallbackClient
	subSeq       uint64

	triggerMu        sync.Mutex
	triggerCallbacks map[string]*GameMenuTriggerCallbackClient
}

func (s *GameMenuModuleRPCServer) Name(_ *Empty, resp *GameMenuModuleNameResp) error {
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

func (s *GameMenuModuleRPCServer) RegisterMenuEntry(args *GameMenuRegisterEntryArgs, resp *GameMenuRegisterEntryResp) error {
	if resp != nil {
		resp.EntryID = ""
	}
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	entry := fromGameMenuEntryWire(args.Entry)

	cb := (*GameMenuTriggerCallbackClient)(nil)
	if args.TriggerBrokerID != 0 && s.broker != nil {
		conn, err := s.broker.Dial(args.TriggerBrokerID)
		if err != nil {
			return err
		}
		cb = &GameMenuTriggerCallbackClient{c: rpc.NewClient(conn)}
		entry.OnTrigger = func(chat *api.ChatMsg) {
			_ = cb.OnTrigger(chat)
		}
	}

	entryID, err := s.Impl.RegisterGameMenuEntry(&entry)
	if err != nil {
		if cb != nil {
			_ = cb.Close()
		}
		return err
	}
	if cb != nil {
		s.triggerMu.Lock()
		if s.triggerCallbacks == nil {
			s.triggerCallbacks = make(map[string]*GameMenuTriggerCallbackClient)
		}
		old := s.triggerCallbacks[entryID]
		s.triggerCallbacks[entryID] = cb
		s.triggerMu.Unlock()
		if old != nil {
			_ = old.Close()
		}
	}
	if resp != nil {
		resp.EntryID = entryID
	}
	return nil
}

func (s *GameMenuModuleRPCServer) RemoveMenuEntry(args *GameMenuRemoveArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	s.Impl.RemoveMenuEntry(args.EntryID)
	s.triggerMu.Lock()
	cb := (*GameMenuTriggerCallbackClient)(nil)
	if s.triggerCallbacks != nil {
		cb = s.triggerCallbacks[args.EntryID]
		delete(s.triggerCallbacks, args.EntryID)
	}
	s.triggerMu.Unlock()
	if cb != nil {
		_ = cb.Close()
	}
	return nil
}

func (s *GameMenuModuleRPCServer) SubscribeEntries(args *GameMenuSubscribeArgs, resp *GameMenuSubscribeResp) error {
	if s == nil || s.Impl == nil || s.broker == nil || args == nil || resp == nil {
		return nil
	}
	if args.CallbackBrokerID == 0 {
		return errors.New("GameMenuModuleRPCServer.SubscribeEntries: callback broker id is 0")
	}

	conn, err := s.broker.Dial(args.CallbackBrokerID)
	if err != nil {
		return err
	}
	cb := &GameMenuEntryCallbackClient{c: rpc.NewClient(conn)}

	ctx, cancel := context.WithCancel(context.Background())
	ch, stop, err := s.Impl.SubscribeEntries(ctx)
	if err != nil {
		cancel()
		cb.Stop()
		_ = cb.Close()
		return err
	}
	if stop == nil {
		stop = func() {}
	}

	subID := fmt.Sprintf("sub:%d", atomic.AddUint64(&s.subSeq, 1))
	s.mu.Lock()
	if s.subs == nil {
		s.subs = make(map[string]context.CancelFunc)
	}
	if s.subCallbacks == nil {
		s.subCallbacks = make(map[string]*GameMenuEntryCallbackClient)
	}
	s.subs[subID] = func() {
		stop()
		cancel()
	}
	s.subCallbacks[subID] = cb
	s.mu.Unlock()

	go func() {
		defer func() {
			s.mu.Lock()
			if s.subs != nil {
				delete(s.subs, subID)
			}
			if s.subCallbacks != nil {
				delete(s.subCallbacks, subID)
			}
			s.mu.Unlock()
			cb.Stop()
			_ = cb.Close()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case info, ok := <-ch:
				if !ok {
					return
				}
				if info == nil {
					continue
				}
				_ = cb.OnEntry(info)
			}
		}
	}()

	resp.SubID = subID
	return nil
}

func (s *GameMenuModuleRPCServer) UnsubscribeEntries(args *GameMenuUnsubscribeArgs, resp *BoolResp) error {
	if s == nil || args == nil {
		return nil
	}
	if args.SubID == "" {
		return nil
	}

	s.mu.Lock()
	cancel := context.CancelFunc(nil)
	if s.subs != nil {
		cancel = s.subs[args.SubID]
		delete(s.subs, args.SubID)
	}
	cb := (*GameMenuEntryCallbackClient)(nil)
	if s.subCallbacks != nil {
		cb = s.subCallbacks[args.SubID]
		delete(s.subCallbacks, args.SubID)
	}
	s.mu.Unlock()

	ok := false
	if cancel != nil {
		cancel()
		ok = true
	}
	if cb != nil {
		cb.Stop()
		_ = cb.Close()
	}
	if resp != nil {
		resp.OK = ok
	}
	return nil
}

func (s *GameMenuModuleRPCServer) TriggerEntry(args *GameMenuTriggerArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	s.Impl.TriggerEntry(args.EntryID, args.Chat)
	return nil
}

type gameMenuModuleRPCClient struct {
	c      *rpc.Client
	broker *plugin.MuxBroker
	mu     sync.Mutex
}

func newGameMenuModuleRPCClient(conn net.Conn, broker *plugin.MuxBroker) api.GameMenuModule {
	if conn == nil {
		return nil
	}
	return &gameMenuModuleRPCClient{
		c:      rpc.NewClient(conn),
		broker: broker,
	}
}

func (c *gameMenuModuleRPCClient) Name() string { return api.NameGameMenuModule }

func (c *gameMenuModuleRPCClient) RegisterGameMenuEntry(entry *api.GameMenuEntry) (string, error) {
	if c == nil || c.c == nil {
		return "", errors.New("gameMenuModuleRPCClient.RegisterMenuEntry: client is not initialised")
	}
	if entry == nil {
		return "", errors.New("gameMenuModuleRPCClient.RegisterMenuEntry: entry is nil")
	}

	var cbID uint32
	if entry.OnTrigger != nil && c.broker != nil {
		cbID = c.broker.NextId()
		go acceptAndServeMuxBroker(c.broker, cbID, &GameMenuTriggerCallbackServer{handler: entry.OnTrigger})
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	var resp GameMenuRegisterEntryResp
	err := c.c.Call("Plugin.RegisterMenuEntry", &GameMenuRegisterEntryArgs{Entry: toGameMenuEntryWire(entry), TriggerBrokerID: cbID}, &resp)
	if err != nil {
		return "", err
	}
	return resp.EntryID, nil
}

func (c *gameMenuModuleRPCClient) RemoveMenuEntry(entryID string) {
	if c == nil || c.c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.c.Call("Plugin.RemoveMenuEntry", &GameMenuRemoveArgs{EntryID: entryID}, &Empty{})
}

func (c *gameMenuModuleRPCClient) SubscribeEntries(ctx context.Context) (<-chan *api.GameMenuEntryInfo, func(), error) {
	if c == nil || c.c == nil || c.broker == nil {
		return nil, nil, errors.New("gameMenuModuleRPCClient.SubscribeEntries: client is not initialised")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	out := make(chan *api.GameMenuEntryInfo, 64)
	cbID := c.broker.NextId()
	cbSrv := &GameMenuEntryCallbackServer{ch: out}
	go acceptAndServeMuxBroker(c.broker, cbID, cbSrv)

	c.mu.Lock()
	var resp GameMenuSubscribeResp
	err := c.c.Call("Plugin.SubscribeEntries", &GameMenuSubscribeArgs{CallbackBrokerID: cbID}, &resp)
	c.mu.Unlock()
	if err != nil {
		_ = cbSrv.Stop(&Empty{}, &Empty{})
		return nil, nil, err
	}
	if resp.SubID == "" {
		_ = cbSrv.Stop(&Empty{}, &Empty{})
		return nil, nil, errors.New("gameMenuModuleRPCClient.SubscribeEntries: empty sub id")
	}

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			c.mu.Lock()
			_ = c.c.Call("Plugin.UnsubscribeEntries", &GameMenuUnsubscribeArgs{SubID: resp.SubID}, &BoolResp{})
			c.mu.Unlock()
			_ = cbSrv.Stop(&Empty{}, &Empty{})
		})
	}
	go func() {
		<-ctx.Done()
		cancel()
	}()
	return out, cancel, nil
}

func (c *gameMenuModuleRPCClient) TriggerEntry(entryID string, chat *api.ChatMsg) {
	if c == nil || c.c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.c.Call("Plugin.TriggerEntry", &GameMenuTriggerArgs{EntryID: entryID, Chat: sanitizeChatMsgForRPC(chat)}, &Empty{})
}

var _ api.GameMenuModule = (*gameMenuModuleRPCClient)(nil)

func nilSafeChatMsg(args *GameMenuTriggerEvent) *api.ChatMsg {
	if args == nil {
		return nil
	}
	return args.Chat
}

func sanitizeChatMsgForRPC(chat *api.ChatMsg) *api.ChatMsg {
	if chat == nil {
		return nil
	}
	out := *chat
	out.Msg = append([]string(nil), chat.Msg...)
	out.RawParameters = append([]string(nil), chat.RawParameters...)
	out.UD = chat.UD
	out.UD.Msg = append([]string(nil), chat.UD.Msg...)
	out.UD.RawParameters = append([]string(nil), chat.UD.RawParameters...)
	out.UD.Aux = nil
	return &out
}
