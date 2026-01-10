package protocol

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"sync"
	"sync/atomic"

	"github.com/hashicorp/go-plugin"

	"github.com/Yeah114/tempest-plugin-sdk/api"
)

type TerminalMenuModuleNameResp struct {
	Name string
}

// TerminalMenuEntryWire is the RPC-safe representation of api.TerminalMenuEntry.
// NOTE: api.TerminalMenuEntry contains a func field (OnTrigger) which is not gob-encodable,
// so we must never send that type over net/rpc directly (even if OnTrigger is nil).
type TerminalMenuEntryWire struct {
	Triggers     []string
	ArgumentHint string
	Usage        string
}

func toTerminalMenuEntryWire(entry *api.TerminalMenuEntry) TerminalMenuEntryWire {
	if entry == nil {
		return TerminalMenuEntryWire{}
	}
	return TerminalMenuEntryWire{
		Triggers:     append([]string(nil), entry.Triggers...),
		ArgumentHint: entry.ArgumentHint,
		Usage:        entry.Usage,
	}
}

func fromTerminalMenuEntryWire(w TerminalMenuEntryWire) api.TerminalMenuEntry {
	return api.TerminalMenuEntry{
		Triggers:     append([]string(nil), w.Triggers...),
		ArgumentHint: w.ArgumentHint,
		Usage:        w.Usage,
	}
}

type TerminalMenuRegisterEntryArgs struct {
	EntryID         string
	Entry           TerminalMenuEntryWire
	TriggerBrokerID uint32
}

type TerminalMenuTriggerEvent struct {
	Args []string
}

type terminalMenuTriggerCallbackServer struct {
	handler func([]string)
}

func (s *terminalMenuTriggerCallbackServer) OnTrigger(args *TerminalMenuTriggerEvent, _ *Empty) error {
	if s == nil || s.handler == nil {
		return nil
	}
	var a []string
	if args != nil && args.Args != nil {
		a = append([]string(nil), args.Args...)
	}
	s.handler(a)
	return nil
}

type terminalMenuTriggerCallbackClient struct {
	c  *rpc.Client
	mu sync.Mutex
}

func (c *terminalMenuTriggerCallbackClient) Close() error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Close()
}

func (c *terminalMenuTriggerCallbackClient) OnTrigger(args []string) error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.OnTrigger", &TerminalMenuTriggerEvent{Args: append([]string(nil), args...)}, &Empty{})
}

type TerminalMenuAddEntryEvent struct {
	EntryID string
	Entry   TerminalMenuEntryWire
}

type terminalMenuAddEntryCallbackServer struct {
	client  *terminalMenuModuleRPCClient
	handler func(*api.TerminalMenuEntry)
}

func (s *terminalMenuAddEntryCallbackServer) OnEntry(args *TerminalMenuAddEntryEvent, _ *Empty) error {
	if s == nil || s.handler == nil || args == nil {
		return nil
	}
	entry := fromTerminalMenuEntryWire(args.Entry)
	entryID := args.EntryID
	if entryID != "" && s.client != nil {
		entry.OnTrigger = func(a []string) {
			_ = s.client.TriggerMenuEntry(entryID, a)
		}
	}
	s.handler(&entry)
	return nil
}

type terminalMenuAddEntryCallbackClient struct {
	c  *rpc.Client
	mu sync.Mutex
}

func (c *terminalMenuAddEntryCallbackClient) Close() error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Close()
}

func (c *terminalMenuAddEntryCallbackClient) OnEntry(entryID string, entry *api.TerminalMenuEntry) error {
	if c == nil || c.c == nil || entry == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.OnEntry", &TerminalMenuAddEntryEvent{EntryID: entryID, Entry: toTerminalMenuEntryWire(entry)}, &Empty{})
}

type TerminalMenuLineEvent struct {
	Line string
}

type terminalMenuLineCallbackServer struct {
	handler func(string)
}

func (s *terminalMenuLineCallbackServer) OnLine(args *TerminalMenuLineEvent, _ *Empty) error {
	if s == nil || s.handler == nil {
		return nil
	}
	line := ""
	if args != nil {
		line = args.Line
	}
	s.handler(line)
	return nil
}

type terminalMenuLineCallbackClient struct {
	c  *rpc.Client
	mu sync.Mutex
}

func (c *terminalMenuLineCallbackClient) Close() error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Close()
}

func (c *terminalMenuLineCallbackClient) OnLine(line string) error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.OnLine", &TerminalMenuLineEvent{Line: line}, &Empty{})
}

type TerminalMenuPopEvent struct{}

type terminalMenuPopCallbackServer struct {
	handler func(struct{})
}

func (s *terminalMenuPopCallbackServer) OnPop(_ *TerminalMenuPopEvent, _ *Empty) error {
	if s == nil || s.handler == nil {
		return nil
	}
	s.handler(struct{}{})
	return nil
}

type terminalMenuPopCallbackClient struct {
	c  *rpc.Client
	mu sync.Mutex
}

func (c *terminalMenuPopCallbackClient) Close() error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Close()
}

func (c *terminalMenuPopCallbackClient) OnPop() error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.OnPop", &TerminalMenuPopEvent{}, &Empty{})
}

type TerminalMenuListenerResp struct {
	ListenerID string
}

type TerminalMenuListenerArgs struct {
	ListenerID string
}

type TerminalMenuRegisterAddEntryArgs struct {
	CallbackBrokerID uint32
}

type TerminalMenuRegisterTerminalCallArgs struct {
	CallbackBrokerID uint32
}

type TerminalMenuRegisterPopArgs struct {
	CallbackBrokerID uint32
}

type TerminalMenuTriggerEntryArgs struct {
	EntryID string
	Args    []string
}

type TerminalMenuRemoveEntryArgs struct {
	EntryID string
}

type TerminalMenuModuleRPCServer struct {
	Impl   api.TerminalMenuModule
	broker *plugin.MuxBroker

	mu            sync.Mutex
	entrySeq      uint64
	entryTriggers map[string]func([]string)

	registeredEntries map[string]*api.TerminalMenuEntry
}

func (s *TerminalMenuModuleRPCServer) Name(_ *Empty, resp *TerminalMenuModuleNameResp) error {
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

func (s *TerminalMenuModuleRPCServer) RegisterMenuEntry(args *TerminalMenuRegisterEntryArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || s.broker == nil || args == nil {
		return nil
	}
	if args.EntryID == "" {
		return errors.New("TerminalMenuModuleRPCServer.RegisterMenuEntry: entry id is empty")
	}
	if args.TriggerBrokerID == 0 {
		return errors.New("TerminalMenuModuleRPCServer.RegisterMenuEntry: trigger broker id is 0")
	}

	conn, err := s.broker.Dial(args.TriggerBrokerID)
	if err != nil {
		return err
	}
	cb := &terminalMenuTriggerCallbackClient{c: rpc.NewClient(conn)}

	entry := fromTerminalMenuEntryWire(args.Entry)
	entry.OnTrigger = func(a []string) {
		_ = cb.OnTrigger(a)
	}
	if err := s.Impl.RegisterTerminalMenuEntry(&entry); err != nil {
		return err
	}

	s.mu.Lock()
	if s.registeredEntries == nil {
		s.registeredEntries = make(map[string]*api.TerminalMenuEntry)
	}
	s.registeredEntries[args.EntryID] = &entry
	s.mu.Unlock()
	return nil
}

func (s *TerminalMenuModuleRPCServer) PublishTerminalCall(args *TerminalMenuLineEvent, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	s.Impl.PublishTerminalCall(args.Line)
	return nil
}

func (s *TerminalMenuModuleRPCServer) PublishPopBackendMenu(_ *Empty, _ *Empty) error {
	if s == nil || s.Impl == nil {
		return nil
	}
	s.Impl.PublishPopBackendMenu()
	return nil
}

func (s *TerminalMenuModuleRPCServer) RegisterWhenAddMenuEntry(args *TerminalMenuRegisterAddEntryArgs, resp *TerminalMenuListenerResp) error {
	if resp == nil {
		return nil
	}
	resp.ListenerID = ""
	if s == nil || s.Impl == nil || s.broker == nil || args == nil {
		return nil
	}
	if args.CallbackBrokerID == 0 {
		return errors.New("TerminalMenuModuleRPCServer.RegisterWhenAddMenuEntry: callback broker id is 0")
	}
	conn, err := s.broker.Dial(args.CallbackBrokerID)
	if err != nil {
		return err
	}
	cb := &terminalMenuAddEntryCallbackClient{c: rpc.NewClient(conn)}

	listenerID, err := s.Impl.RegisterWhenAddMenuEntry(func(entry *api.TerminalMenuEntry) {
		if entry == nil {
			return
		}

		entryID := fmt.Sprintf("entry:%d", atomic.AddUint64(&s.entrySeq, 1))
		s.mu.Lock()
		if s.entryTriggers == nil {
			s.entryTriggers = make(map[string]func([]string))
		}
		if entry.OnTrigger != nil {
			s.entryTriggers[entryID] = entry.OnTrigger
		}
		s.mu.Unlock()

		clean := *entry
		clean.OnTrigger = nil
		_ = cb.OnEntry(entryID, &clean)
	})
	if err != nil {
		_ = cb.Close()
		return err
	}
	resp.ListenerID = listenerID
	return nil
}

func (s *TerminalMenuModuleRPCServer) UnregisterWhenAddMenuEntry(args *TerminalMenuListenerArgs, resp *BoolResp) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	ok := s.Impl.UnregisterWhenAddMenuEntry(args.ListenerID)
	if resp != nil {
		resp.OK = ok
	}
	return nil
}

func (s *TerminalMenuModuleRPCServer) RegisterWhenTerminalCall(args *TerminalMenuRegisterTerminalCallArgs, resp *TerminalMenuListenerResp) error {
	if resp == nil {
		return nil
	}
	resp.ListenerID = ""
	if s == nil || s.Impl == nil || s.broker == nil || args == nil {
		return nil
	}
	if args.CallbackBrokerID == 0 {
		return errors.New("TerminalMenuModuleRPCServer.RegisterWhenTerminalCall: callback broker id is 0")
	}
	conn, err := s.broker.Dial(args.CallbackBrokerID)
	if err != nil {
		return err
	}
	cb := &terminalMenuLineCallbackClient{c: rpc.NewClient(conn)}

	listenerID, err := s.Impl.RegisterWhenTerminalCall(func(line string) {
		_ = cb.OnLine(line)
	})
	if err != nil {
		_ = cb.Close()
		return err
	}
	resp.ListenerID = listenerID
	return nil
}

func (s *TerminalMenuModuleRPCServer) UnregisterWhenTerminalCall(args *TerminalMenuListenerArgs, resp *BoolResp) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	ok := s.Impl.UnregisterWhenTerminalCall(args.ListenerID)
	if resp != nil {
		resp.OK = ok
	}
	return nil
}

func (s *TerminalMenuModuleRPCServer) RegisterWhenPopBackendMenu(args *TerminalMenuRegisterPopArgs, resp *TerminalMenuListenerResp) error {
	if resp == nil {
		return nil
	}
	resp.ListenerID = ""
	if s == nil || s.Impl == nil || s.broker == nil || args == nil {
		return nil
	}
	if args.CallbackBrokerID == 0 {
		return errors.New("TerminalMenuModuleRPCServer.RegisterWhenPopBackendMenu: callback broker id is 0")
	}
	conn, err := s.broker.Dial(args.CallbackBrokerID)
	if err != nil {
		return err
	}
	cb := &terminalMenuPopCallbackClient{c: rpc.NewClient(conn)}

	listenerID, err := s.Impl.RegisterWhenPopBackendMenu(func(_ struct{}) {
		_ = cb.OnPop()
	})
	if err != nil {
		_ = cb.Close()
		return err
	}
	resp.ListenerID = listenerID
	return nil
}

func (s *TerminalMenuModuleRPCServer) UnregisterWhenPopBackendMenu(args *TerminalMenuListenerArgs, resp *BoolResp) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	ok := s.Impl.UnregisterWhenPopBackendMenu(args.ListenerID)
	if resp != nil {
		resp.OK = ok
	}
	return nil
}

func (s *TerminalMenuModuleRPCServer) TriggerMenuEntry(args *TerminalMenuTriggerEntryArgs, resp *BoolResp) error {
	if s == nil || args == nil || args.EntryID == "" {
		return nil
	}
	s.mu.Lock()
	fn := s.entryTriggers[args.EntryID]
	s.mu.Unlock()
	if fn == nil {
		if resp != nil {
			resp.OK = false
		}
		return nil
	}
	fn(args.Args)
	if resp != nil {
		resp.OK = true
	}
	return nil
}

func (s *TerminalMenuModuleRPCServer) RemoveMenuEntry(args *TerminalMenuRemoveEntryArgs, resp *BoolResp) error {
	if s == nil || s.Impl == nil || args == nil || args.EntryID == "" {
		return nil
	}
	s.mu.Lock()
	entry := s.registeredEntries[args.EntryID]
	delete(s.registeredEntries, args.EntryID)
	s.mu.Unlock()
	if entry == nil {
		if resp != nil {
			resp.OK = false
		}
		return nil
	}
	ok := s.Impl.RemoveTerminalMenuEntry(entry)
	if resp != nil {
		resp.OK = ok
	}
	return nil
}

type terminalMenuModuleRPCClient struct {
	c      *rpc.Client
	broker *plugin.MuxBroker
	mu     sync.Mutex

	entrySeq uint64
	entryMu  sync.Mutex
	entries  map[*api.TerminalMenuEntry]string
}

func newTerminalMenuModuleRPCClient(conn net.Conn, broker *plugin.MuxBroker) api.TerminalMenuModule {
	if conn == nil {
		return nil
	}
	return &terminalMenuModuleRPCClient{
		c:      rpc.NewClient(conn),
		broker: broker,
	}
}

func (c *terminalMenuModuleRPCClient) Name() string { return api.NameTerminalMenuModule }

func (c *terminalMenuModuleRPCClient) RegisterTerminalMenuEntry(entry *api.TerminalMenuEntry) error {
	if c == nil || c.c == nil || c.broker == nil {
		return errors.New("terminalMenuModuleRPCClient.RegisterMenuEntry: client is not initialised")
	}
	if entry == nil {
		return errors.New("terminalMenuModuleRPCClient.RegisterMenuEntry: entry is nil")
	}
	if entry.OnTrigger == nil {
		return errors.New("terminalMenuModuleRPCClient.RegisterMenuEntry: entry.OnTrigger is nil")
	}

	cbID := c.broker.NextId()
	go acceptAndServeMuxBroker(c.broker, cbID, &terminalMenuTriggerCallbackServer{handler: entry.OnTrigger})

	entryID := fmt.Sprintf("entry:%d", atomic.AddUint64(&c.entrySeq, 1))
	c.entryMu.Lock()
	if c.entries == nil {
		c.entries = make(map[*api.TerminalMenuEntry]string)
	}
	c.entries[entry] = entryID
	c.entryMu.Unlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.RegisterMenuEntry", &TerminalMenuRegisterEntryArgs{EntryID: entryID, Entry: toTerminalMenuEntryWire(entry), TriggerBrokerID: cbID}, &Empty{})
}

func (c *terminalMenuModuleRPCClient) RemoveTerminalMenuEntry(entry *api.TerminalMenuEntry) bool {
	if c == nil || c.c == nil || entry == nil {
		return false
	}
	c.entryMu.Lock()
	entryID := ""
	if c.entries != nil {
		entryID = c.entries[entry]
	}
	c.entryMu.Unlock()
	if entryID == "" {
		return false
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	var resp BoolResp
	if err := c.c.Call("Plugin.RemoveMenuEntry", &TerminalMenuRemoveEntryArgs{EntryID: entryID}, &resp); err != nil {
		return false
	}
	if resp.OK {
		c.entryMu.Lock()
		delete(c.entries, entry)
		c.entryMu.Unlock()
	}
	return resp.OK
}

func (c *terminalMenuModuleRPCClient) PublishTerminalCall(line string) {
	if c == nil || c.c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.c.Call("Plugin.PublishTerminalCall", &TerminalMenuLineEvent{Line: line}, &Empty{})
}

func (c *terminalMenuModuleRPCClient) PublishPopBackendMenu() {
	if c == nil || c.c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.c.Call("Plugin.PublishPopBackendMenu", &Empty{}, &Empty{})
}

func (c *terminalMenuModuleRPCClient) RegisterWhenAddMenuEntry(handler func(*api.TerminalMenuEntry)) (string, error) {
	if c == nil || c.c == nil || c.broker == nil {
		return "", errors.New("terminalMenuModuleRPCClient.RegisterWhenAddMenuEntry: client is not initialised")
	}
	if handler == nil {
		return "", errors.New("terminalMenuModuleRPCClient.RegisterWhenAddMenuEntry: handler is nil")
	}

	cbID := c.broker.NextId()
	go acceptAndServeMuxBroker(c.broker, cbID, &terminalMenuAddEntryCallbackServer{client: c, handler: handler})

	c.mu.Lock()
	defer c.mu.Unlock()
	var resp TerminalMenuListenerResp
	if err := c.c.Call("Plugin.RegisterWhenAddMenuEntry", &TerminalMenuRegisterAddEntryArgs{CallbackBrokerID: cbID}, &resp); err != nil {
		return "", err
	}
	return resp.ListenerID, nil
}

func (c *terminalMenuModuleRPCClient) UnregisterWhenAddMenuEntry(listenerID string) bool {
	if c == nil || c.c == nil {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp BoolResp
	if err := c.c.Call("Plugin.UnregisterWhenAddMenuEntry", &TerminalMenuListenerArgs{ListenerID: listenerID}, &resp); err != nil {
		return false
	}
	return resp.OK
}

func (c *terminalMenuModuleRPCClient) RegisterWhenTerminalCall(handler func(string)) (string, error) {
	if c == nil || c.c == nil || c.broker == nil {
		return "", errors.New("terminalMenuModuleRPCClient.RegisterWhenTerminalCall: client is not initialised")
	}
	if handler == nil {
		return "", errors.New("terminalMenuModuleRPCClient.RegisterWhenTerminalCall: handler is nil")
	}

	cbID := c.broker.NextId()
	go acceptAndServeMuxBroker(c.broker, cbID, &terminalMenuLineCallbackServer{handler: handler})

	c.mu.Lock()
	defer c.mu.Unlock()
	var resp TerminalMenuListenerResp
	if err := c.c.Call("Plugin.RegisterWhenTerminalCall", &TerminalMenuRegisterTerminalCallArgs{CallbackBrokerID: cbID}, &resp); err != nil {
		return "", err
	}
	return resp.ListenerID, nil
}

func (c *terminalMenuModuleRPCClient) UnregisterWhenTerminalCall(listenerID string) bool {
	if c == nil || c.c == nil {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp BoolResp
	if err := c.c.Call("Plugin.UnregisterWhenTerminalCall", &TerminalMenuListenerArgs{ListenerID: listenerID}, &resp); err != nil {
		return false
	}
	return resp.OK
}

func (c *terminalMenuModuleRPCClient) RegisterWhenPopBackendMenu(handler func(struct{})) (string, error) {
	if c == nil || c.c == nil || c.broker == nil {
		return "", errors.New("terminalMenuModuleRPCClient.RegisterWhenPopBackendMenu: client is not initialised")
	}
	if handler == nil {
		return "", errors.New("terminalMenuModuleRPCClient.RegisterWhenPopBackendMenu: handler is nil")
	}

	cbID := c.broker.NextId()
	go acceptAndServeMuxBroker(c.broker, cbID, &terminalMenuPopCallbackServer{handler: handler})

	c.mu.Lock()
	defer c.mu.Unlock()
	var resp TerminalMenuListenerResp
	if err := c.c.Call("Plugin.RegisterWhenPopBackendMenu", &TerminalMenuRegisterPopArgs{CallbackBrokerID: cbID}, &resp); err != nil {
		return "", err
	}
	return resp.ListenerID, nil
}

func (c *terminalMenuModuleRPCClient) UnregisterWhenPopBackendMenu(listenerID string) bool {
	if c == nil || c.c == nil {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp BoolResp
	if err := c.c.Call("Plugin.UnregisterWhenPopBackendMenu", &TerminalMenuListenerArgs{ListenerID: listenerID}, &resp); err != nil {
		return false
	}
	return resp.OK
}

func (c *terminalMenuModuleRPCClient) TriggerMenuEntry(entryID string, args []string) error {
	if c == nil || c.c == nil {
		return errors.New("terminalMenuModuleRPCClient.TriggerMenuEntry: client is not initialised")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.TriggerMenuEntry", &TerminalMenuTriggerEntryArgs{EntryID: entryID, Args: append([]string(nil), args...)}, &BoolResp{})
}
