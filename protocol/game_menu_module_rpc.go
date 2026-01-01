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

	"github.com/Yeah114/tempest-plugin-sdk/api"
)

type GameMenuModuleNameResp struct {
	Name string
}

type gameMenuEntryEvent struct {
	Entry api.GameMenuEntry
}

type gameMenuEntryCallbackServer struct {
	mu     sync.Mutex
	closed bool
	ch     chan<- *api.GameMenuEntry
}

func (s *gameMenuEntryCallbackServer) OnEntry(args *gameMenuEntryEvent, _ *Empty) error {
	if s == nil {
		return nil
	}
	var entry *api.GameMenuEntry
	if args != nil {
		e := args.Entry
		entry = &e
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.ch == nil || entry == nil {
		return nil
	}
	select {
	case s.ch <- entry:
	default:
	}
	return nil
}

func (s *gameMenuEntryCallbackServer) Stop(_ *Empty, _ *Empty) error {
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

type gameMenuEntryCallbackClient struct {
	c  *rpc.Client
	mu sync.Mutex
}

func (c *gameMenuEntryCallbackClient) Close() error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Close()
}

func (c *gameMenuEntryCallbackClient) OnEntry(entry *api.GameMenuEntry) error {
	if c == nil || c.c == nil || entry == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.OnEntry", &gameMenuEntryEvent{Entry: *entry}, &Empty{})
}

func (c *gameMenuEntryCallbackClient) Stop() {
	if c == nil || c.c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.c.Call("Plugin.Stop", &Empty{}, &Empty{})
}

type gameMenuTriggerEvent struct {
	ChatJSON []byte
}

type gameMenuTriggerCallbackServer struct {
	handler func([]byte)
}

func (s *gameMenuTriggerCallbackServer) OnTrigger(args *gameMenuTriggerEvent, _ *Empty) error {
	if s == nil || s.handler == nil {
		return nil
	}
	var payload []byte
	if args != nil && args.ChatJSON != nil {
		payload = append([]byte(nil), args.ChatJSON...)
	}
	s.handler(payload)
	return nil
}

type gameMenuTriggerCallbackClient struct {
	c  *rpc.Client
	mu sync.Mutex
}

func (c *gameMenuTriggerCallbackClient) Close() error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Close()
}

func (c *gameMenuTriggerCallbackClient) OnTrigger(chatJSON []byte) error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.OnTrigger", &gameMenuTriggerEvent{ChatJSON: append([]byte(nil), chatJSON...)}, &Empty{})
}

type GameMenuRegisterArgs struct {
	Entry *api.GameMenuEntry
}

type GameMenuRemoveArgs struct {
	ID string
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

type GameMenuRegisterTriggerArgs struct {
	ID               string
	CallbackBrokerID uint32
}

type GameMenuTriggerArgs struct {
	ID      string
	ChatJSON []byte
}

type GameMenuModuleRPCServer struct {
	Impl   api.GameMenuModule
	broker *plugin.MuxBroker

	mu           sync.Mutex
	subs         map[string]context.CancelFunc
	subCallbacks map[string]*gameMenuEntryCallbackClient
	subSeq       uint64

	triggerMu        sync.Mutex
	triggerCallbacks map[string]*gameMenuTriggerCallbackClient
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

func (s *GameMenuModuleRPCServer) RegisterMenuEntry(args *GameMenuRegisterArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	return s.Impl.RegisterMenuEntry(args.Entry)
}

func (s *GameMenuModuleRPCServer) RemoveMenuEntry(args *GameMenuRemoveArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	s.Impl.RemoveMenuEntry(args.ID)
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
	cb := &gameMenuEntryCallbackClient{c: rpc.NewClient(conn)}

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
		s.subCallbacks = make(map[string]*gameMenuEntryCallbackClient)
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
			case entry, ok := <-ch:
				if !ok {
					return
				}
				_ = cb.OnEntry(entry)
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
	cb := (*gameMenuEntryCallbackClient)(nil)
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

func (s *GameMenuModuleRPCServer) RegisterTriggerHandler(args *GameMenuRegisterTriggerArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || s.broker == nil || args == nil {
		return nil
	}
	if args.ID == "" {
		return errors.New("GameMenuModuleRPCServer.RegisterTriggerHandler: id is empty")
	}
	if args.CallbackBrokerID == 0 {
		return errors.New("GameMenuModuleRPCServer.RegisterTriggerHandler: callback broker id is 0")
	}

	conn, err := s.broker.Dial(args.CallbackBrokerID)
	if err != nil {
		return err
	}
	cb := &gameMenuTriggerCallbackClient{c: rpc.NewClient(conn)}

	if err := s.Impl.RegisterTriggerHandler(args.ID, func(chatJSON []byte) {
		_ = cb.OnTrigger(chatJSON)
	}); err != nil {
		_ = cb.Close()
		return err
	}

	s.triggerMu.Lock()
	if s.triggerCallbacks == nil {
		s.triggerCallbacks = make(map[string]*gameMenuTriggerCallbackClient)
	}
	old := s.triggerCallbacks[args.ID]
	s.triggerCallbacks[args.ID] = cb
	s.triggerMu.Unlock()
	if old != nil {
		_ = old.Close()
	}
	return nil
}

func (s *GameMenuModuleRPCServer) RemoveTriggerHandler(args *GameMenuRemoveArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	if args.ID == "" {
		return nil
	}
	s.Impl.RemoveTriggerHandler(args.ID)
	s.triggerMu.Lock()
	cb := (*gameMenuTriggerCallbackClient)(nil)
	if s.triggerCallbacks != nil {
		cb = s.triggerCallbacks[args.ID]
		delete(s.triggerCallbacks, args.ID)
	}
	s.triggerMu.Unlock()
	if cb != nil {
		_ = cb.Close()
	}
	return nil
}

func (s *GameMenuModuleRPCServer) TriggerEntry(args *GameMenuTriggerArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	s.Impl.TriggerEntry(args.ID, args.ChatJSON)
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

func (c *gameMenuModuleRPCClient) RegisterMenuEntry(entry *api.GameMenuEntry) error {
	if c == nil || c.c == nil {
		return errors.New("gameMenuModuleRPCClient.RegisterMenuEntry: client is not initialised")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.RegisterMenuEntry", &GameMenuRegisterArgs{Entry: entry}, &Empty{})
}

func (c *gameMenuModuleRPCClient) RemoveMenuEntry(id string) {
	if c == nil || c.c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.c.Call("Plugin.RemoveMenuEntry", &GameMenuRemoveArgs{ID: id}, &Empty{})
}

func (c *gameMenuModuleRPCClient) SubscribeEntries(ctx context.Context) (<-chan *api.GameMenuEntry, func(), error) {
	if c == nil || c.c == nil || c.broker == nil {
		return nil, nil, errors.New("gameMenuModuleRPCClient.SubscribeEntries: client is not initialised")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	out := make(chan *api.GameMenuEntry, 64)
	cbID := c.broker.NextId()
	cbSrv := &gameMenuEntryCallbackServer{ch: out}
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

func (c *gameMenuModuleRPCClient) RegisterTriggerHandler(id string, handler func(chatJSON []byte)) error {
	if c == nil || c.c == nil || c.broker == nil {
		return errors.New("gameMenuModuleRPCClient.RegisterTriggerHandler: client is not initialised")
	}
	if id == "" {
		return errors.New("gameMenuModuleRPCClient.RegisterTriggerHandler: id is empty")
	}
	if handler == nil {
		return errors.New("gameMenuModuleRPCClient.RegisterTriggerHandler: handler is nil")
	}

	cbID := c.broker.NextId()
	go acceptAndServeMuxBroker(c.broker, cbID, &gameMenuTriggerCallbackServer{handler: handler})

	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.RegisterTriggerHandler", &GameMenuRegisterTriggerArgs{ID: id, CallbackBrokerID: cbID}, &Empty{})
}

func (c *gameMenuModuleRPCClient) RemoveTriggerHandler(id string) {
	if c == nil || c.c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.c.Call("Plugin.RemoveTriggerHandler", &GameMenuRemoveArgs{ID: id}, &Empty{})
}

func (c *gameMenuModuleRPCClient) TriggerEntry(id string, chatJSON []byte) {
	if c == nil || c.c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.c.Call("Plugin.TriggerEntry", &GameMenuTriggerArgs{ID: id, ChatJSON: chatJSON}, &Empty{})
}
