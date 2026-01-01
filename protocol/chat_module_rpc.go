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

type ChatMsgEvent struct {
	Event api.ChatMsg
}

type chatMsgCallbackServer struct {
	handler func(*api.ChatMsg)
}

func (s *chatMsgCallbackServer) OnChatMsg(args *ChatMsgEvent, _ *Empty) error {
	if s == nil || s.handler == nil || args == nil {
		return nil
	}
	event := args.Event
	s.handler(&event)
	return nil
}

type chatMsgCallbackClient struct {
	c  *rpc.Client
	mu sync.Mutex
}

func (c *chatMsgCallbackClient) Close() error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Close()
}

func (c *chatMsgCallbackClient) OnChatMsg(event *api.ChatMsg) error {
	if c == nil || c.c == nil || event == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	safe := *event
	safe.UD.Aux = nil
	return c.c.Call("Plugin.OnChatMsg", &ChatMsgEvent{Event: safe}, &Empty{})
}

type ChatModuleRegisterArgs struct {
	CallbackBrokerID uint32
	SenderName       string
}

type ChatModuleRegisterResp struct {
	ListenerID string
}

type ChatModuleUnregisterArgs struct {
	ListenerID string
}

type BoolResp struct {
	OK bool
}

type ChatModuleRPCServer struct {
	Impl   api.ChatModule
	broker *plugin.MuxBroker

	mu        sync.Mutex
	callbacks map[string]*chatMsgCallbackClient

	interceptMu       sync.Mutex
	interceptCancels  map[string]func()
	interceptCallbacks map[string]*chatMsgCallbackClient
	interceptSeq      uint64
}

type ChatModuleNameResp struct {
	Name string
}

func (s *ChatModuleRPCServer) Name(_ *Empty, resp *ChatModuleNameResp) error {
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

func (s *ChatModuleRPCServer) RegisterWhenChatMsg(args *ChatModuleRegisterArgs, resp *ChatModuleRegisterResp) error {
	if s == nil || s.Impl == nil || s.broker == nil || args == nil || resp == nil {
		return nil
	}
	if args.CallbackBrokerID == 0 {
		return errors.New("ChatModuleRPCServer.RegisterWhenChatMsg: callback broker id is 0")
	}
	conn, err := s.broker.Dial(args.CallbackBrokerID)
	if err != nil {
		return err
	}
	cb := &chatMsgCallbackClient{c: rpc.NewClient(conn)}

	listenerID, err := s.Impl.RegisterWhenChatMsg(func(event *api.ChatMsg) {
		_ = cb.OnChatMsg(event)
	})
	if err != nil {
		_ = cb.Close()
		return err
	}

	s.mu.Lock()
	if s.callbacks == nil {
		s.callbacks = make(map[string]*chatMsgCallbackClient)
	}
	s.callbacks[listenerID] = cb
	s.mu.Unlock()

	resp.ListenerID = listenerID
	return nil
}

func (s *ChatModuleRPCServer) UnregisterWhenChatMsg(args *ChatModuleUnregisterArgs, resp *BoolResp) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	if args.ListenerID == "" {
		return nil
	}

	ok := s.Impl.UnregisterWhenChatMsg(args.ListenerID)
	if resp != nil {
		resp.OK = ok
	}

	s.mu.Lock()
	cb := s.callbacks[args.ListenerID]
	delete(s.callbacks, args.ListenerID)
	s.mu.Unlock()
	if cb != nil {
		_ = cb.Close()
	}
	return nil
}

func (s *ChatModuleRPCServer) RegisterWhenReceiveMsgFromSenderNamed(args *ChatModuleRegisterArgs, resp *ChatModuleRegisterResp) error {
	if s == nil || s.Impl == nil || s.broker == nil || args == nil || resp == nil {
		return nil
	}
	if args.SenderName == "" {
		return errors.New("ChatModuleRPCServer.RegisterWhenReceiveMsgFromSenderNamed: sender name is empty")
	}
	if args.CallbackBrokerID == 0 {
		return errors.New("ChatModuleRPCServer.RegisterWhenReceiveMsgFromSenderNamed: callback broker id is 0")
	}

	conn, err := s.broker.Dial(args.CallbackBrokerID)
	if err != nil {
		return err
	}
	cb := &chatMsgCallbackClient{c: rpc.NewClient(conn)}

	listenerID, err := s.Impl.RegisterWhenReceiveMsgFromSenderNamed(args.SenderName, func(event *api.ChatMsg) {
		_ = cb.OnChatMsg(event)
	})

	if err != nil {
		_ = cb.Close()
		return err
	}

	s.mu.Lock()
	if s.callbacks == nil {
		s.callbacks = make(map[string]*chatMsgCallbackClient)
	}
	s.callbacks[listenerID] = cb
	s.mu.Unlock()

	resp.ListenerID = listenerID
	return nil
}

func (s *ChatModuleRPCServer) UnregisterWhenReceiveMsgFromSenderNamed(args *ChatModuleUnregisterArgs, resp *BoolResp) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	if args.ListenerID == "" {
		return nil
	}

	ok := s.Impl.UnregisterWhenReceiveMsgFromSenderNamed(args.ListenerID)
	if resp != nil {
		resp.OK = ok
	}

	s.mu.Lock()
	cb := s.callbacks[args.ListenerID]
	delete(s.callbacks, args.ListenerID)
	s.mu.Unlock()
	if cb != nil {
		_ = cb.Close()
	}
	return nil
}

type ChatModuleInterceptResp struct {
	InterceptID string
}

type ChatModuleCancelInterceptArgs struct {
	InterceptID string
}

func (s *ChatModuleRPCServer) InterceptNextMessage(args *ChatModuleRegisterArgs, resp *ChatModuleInterceptResp) error {
	if s == nil || s.Impl == nil || s.broker == nil || args == nil {
		return nil
	}
	if args.SenderName == "" {
		return errors.New("ChatModuleRPCServer.InterceptNextMessage: sender name is empty")
	}
	if args.CallbackBrokerID == 0 {
		return errors.New("ChatModuleRPCServer.InterceptNextMessage: callback broker id is 0")
	}

	conn, err := s.broker.Dial(args.CallbackBrokerID)
	if err != nil {
		return err
	}
	cb := &chatMsgCallbackClient{c: rpc.NewClient(conn)}

	seq := atomic.AddUint64(&s.interceptSeq, 1)
	interceptID := fmt.Sprintf("intercept:%d", seq)

	cancel, err := s.Impl.InterceptNextMessage(args.SenderName, func(event *api.ChatMsg) {
		_ = cb.OnChatMsg(event)
		_ = cb.Close()
		s.interceptMu.Lock()
		delete(s.interceptCancels, interceptID)
		delete(s.interceptCallbacks, interceptID)
		s.interceptMu.Unlock()
	})
	if err != nil {
		_ = cb.Close()
		return err
	}

	s.interceptMu.Lock()
	if s.interceptCancels == nil {
		s.interceptCancels = make(map[string]func())
	}
	if s.interceptCallbacks == nil {
		s.interceptCallbacks = make(map[string]*chatMsgCallbackClient)
	}
	s.interceptCancels[interceptID] = cancel
	s.interceptCallbacks[interceptID] = cb
	s.interceptMu.Unlock()

	if resp != nil {
		resp.InterceptID = interceptID
	}
	return nil
}

func (s *ChatModuleRPCServer) CancelIntercept(args *ChatModuleCancelInterceptArgs, resp *BoolResp) error {
	if s == nil || args == nil {
		return nil
	}
	if args.InterceptID == "" {
		return nil
	}

	var cancel func()
	var cb *chatMsgCallbackClient

	s.interceptMu.Lock()
	if s.interceptCancels != nil {
		cancel = s.interceptCancels[args.InterceptID]
		delete(s.interceptCancels, args.InterceptID)
	}
	if s.interceptCallbacks != nil {
		cb = s.interceptCallbacks[args.InterceptID]
		delete(s.interceptCallbacks, args.InterceptID)
	}
	s.interceptMu.Unlock()

	ok := false
	if cancel != nil {
		cancel()
		ok = true
	}
	if cb != nil {
		_ = cb.Close()
	}
	if resp != nil {
		resp.OK = ok
	}
	return nil
}

type chatModuleRPCClient struct {
	c      *rpc.Client
	broker *plugin.MuxBroker
	mu     sync.Mutex
}

func newChatModuleRPCClient(conn net.Conn, broker *plugin.MuxBroker) api.ChatModule {
	if conn == nil {
		return nil
	}
	return &chatModuleRPCClient{
		c:      rpc.NewClient(conn),
		broker: broker,
	}
}

func (c *chatModuleRPCClient) Name() string {
	if c == nil || c.c == nil {
		return api.NameChatModule
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	var resp ChatModuleNameResp
	_ = c.c.Call("Plugin.Name", &Empty{}, &resp)
	if resp.Name == "" {
		return api.NameChatModule
	}
	return resp.Name
}

func (c *chatModuleRPCClient) RegisterWhenChatMsg(handler func(event *api.ChatMsg)) (string, error) {
	if c == nil || c.c == nil || c.broker == nil {
		return "", errors.New("chatModuleRPCClient.RegisterWhenChatMsg: client is not initialised")
	}
	if handler == nil {
		return "", errors.New("chatModuleRPCClient.RegisterWhenChatMsg: handler is nil")
	}

	cbID := c.broker.NextId()
	go c.broker.AcceptAndServe(cbID, &chatMsgCallbackServer{handler: handler})

	c.mu.Lock()
	defer c.mu.Unlock()

	var resp ChatModuleRegisterResp
	if err := c.c.Call("Plugin.RegisterWhenChatMsg", &ChatModuleRegisterArgs{CallbackBrokerID: cbID}, &resp); err != nil {
		return "", err
	}
	return resp.ListenerID, nil
}

func (c *chatModuleRPCClient) UnregisterWhenChatMsg(listenerID string) bool {
	if c == nil || c.c == nil {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp BoolResp
	if err := c.c.Call("Plugin.UnregisterWhenChatMsg", &ChatModuleUnregisterArgs{ListenerID: listenerID}, &resp); err != nil {
		return false
	}
	return resp.OK
}

func (c *chatModuleRPCClient) RegisterWhenReceiveMsgFromSenderNamed(name string, handler func(event *api.ChatMsg)) (string, error) {
	if c == nil || c.c == nil || c.broker == nil {
		return "", errors.New("chatModuleRPCClient.RegisterWhenReceiveMsgFromSenderNamed: client is not initialised")
	}
	if handler == nil {
		return "", errors.New("chatModuleRPCClient.RegisterWhenReceiveMsgFromSenderNamed: handler is nil")
	}
	if name == "" {
		return "", errors.New("chatModuleRPCClient.RegisterWhenReceiveMsgFromSenderNamed: name is empty")
	}

	cbID := c.broker.NextId()
	go c.broker.AcceptAndServe(cbID, &chatMsgCallbackServer{handler: handler})

	c.mu.Lock()
	defer c.mu.Unlock()

	var resp ChatModuleRegisterResp
	if err := c.c.Call("Plugin.RegisterWhenReceiveMsgFromSenderNamed", &ChatModuleRegisterArgs{CallbackBrokerID: cbID, SenderName: name}, &resp); err != nil {
		return "", err
	}
	return resp.ListenerID, nil
}

func (c *chatModuleRPCClient) UnregisterWhenReceiveMsgFromSenderNamed(listenerID string) bool {
	if c == nil || c.c == nil {
		return false
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	var resp BoolResp
	if err := c.c.Call("Plugin.UnregisterWhenReceiveMsgFromSenderNamed", &ChatModuleUnregisterArgs{ListenerID: listenerID}, &resp); err != nil {
		return false
	}
	return resp.OK
}

func (c *chatModuleRPCClient) InterceptNextMessage(name string, handler func(*api.ChatMsg)) (func(), error) {
	if c == nil || c.c == nil || c.broker == nil {
		return nil, errors.New("chatModuleRPCClient.InterceptNextMessage: client is not initialised")
	}
	if handler == nil {
		return nil, errors.New("chatModuleRPCClient.InterceptNextMessage: handler is nil")
	}
	if name == "" {
		return nil, errors.New("chatModuleRPCClient.InterceptNextMessage: name is empty")
	}
	cbID := c.broker.NextId()
	go c.broker.AcceptAndServe(cbID, &chatMsgCallbackServer{handler: handler})

	c.mu.Lock()
	var resp ChatModuleInterceptResp
	err := c.c.Call("Plugin.InterceptNextMessage", &ChatModuleRegisterArgs{CallbackBrokerID: cbID, SenderName: name}, &resp)
	c.mu.Unlock()
	if err != nil {
		return nil, err
	}
	if resp.InterceptID == "" {
		return func() {}, nil
	}

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			_ = c.c.Call("Plugin.CancelIntercept", &ChatModuleCancelInterceptArgs{InterceptID: resp.InterceptID}, &BoolResp{})
		})
	}
	return cancel, nil
}
