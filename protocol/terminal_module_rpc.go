package protocol

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/go-plugin"

	"github.com/Yeah114/tempest-plugin-sdk/api"
)

type terminalLineEvent struct {
	Line string
}

type terminalLineCallbackServer struct {
	mu     sync.Mutex
	closed bool
	ch     chan<- string
}

func (s *terminalLineCallbackServer) OnLine(args *terminalLineEvent, _ *Empty) error {
	if s == nil {
		return nil
	}
	line := ""
	if args != nil {
		line = args.Line
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.ch == nil {
		return nil
	}
	select {
	case s.ch <- line:
	default:
	}
	return nil
}

func (s *terminalLineCallbackServer) Stop(_ *Empty, _ *Empty) error {
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

type terminalLineCallbackClient struct {
	c  *rpc.Client
	mu sync.Mutex
}

func (c *terminalLineCallbackClient) Close() error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Close()
}

func (c *terminalLineCallbackClient) OnLine(line string) error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.OnLine", &terminalLineEvent{Line: line}, &Empty{})
}

func (c *terminalLineCallbackClient) Stop() {
	if c == nil || c.c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.c.Call("Plugin.Stop", &Empty{}, &Empty{})
}

type TerminalModuleNameResp struct {
	Name string
}

type TerminalPrintArgs struct {
	Level api.Level
	Scope string
	Msg   string
}

type TerminalRawArgs struct {
	Msg string
}

type TerminalColorTransArgs struct {
	Msg string
}

type TerminalColorTransResp struct {
	Msg string
}

type TerminalSubscribeArgs struct {
	CallbackBrokerID uint32
}

type TerminalSubscribeResp struct {
	SubID string
}

type TerminalUnsubscribeArgs struct {
	SubID string
}

type TerminalInterceptArgs struct {
	CallbackBrokerID uint32
	TimeoutMillis    int64
}

type TerminalInterceptResp struct {
	InterceptID string
}

type TerminalCancelInterceptArgs struct {
	InterceptID string
}

type TerminalModuleRPCServer struct {
	Impl   api.TerminalModule
	broker *plugin.MuxBroker

	mu        sync.Mutex
	subs      map[string]context.CancelFunc
	callbacks map[string]*terminalLineCallbackClient
	seq       uint64

	interceptMu        sync.Mutex
	interceptCancels   map[string]func()
	interceptCallbacks map[string]*terminalLineCallbackClient
	interceptSeq       uint64
}

func (s *TerminalModuleRPCServer) Name(_ *Empty, resp *TerminalModuleNameResp) error {
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

func (s *TerminalModuleRPCServer) Print(args *TerminalPrintArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	s.Impl.Print(args.Level, args.Scope, args.Msg)
	return nil
}

func (s *TerminalModuleRPCServer) Info(args *TerminalRawArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	s.Impl.Info("", args.Msg)
	return nil
}

func (s *TerminalModuleRPCServer) Warn(args *TerminalRawArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	s.Impl.Warn("", args.Msg)
	return nil
}

func (s *TerminalModuleRPCServer) Error(args *TerminalRawArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	s.Impl.Error("", args.Msg)
	return nil
}

func (s *TerminalModuleRPCServer) Success(args *TerminalRawArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	s.Impl.Success("", args.Msg)
	return nil
}

func (s *TerminalModuleRPCServer) Raw(args *TerminalRawArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	s.Impl.Raw(args.Msg)
	return nil
}

func (s *TerminalModuleRPCServer) ColorTransANSI(args *TerminalColorTransArgs, resp *TerminalColorTransResp) error {
	if resp == nil {
		return nil
	}
	resp.Msg = ""
	if s == nil || s.Impl == nil {
		return nil
	}
	msg := ""
	if args != nil {
		msg = args.Msg
	}
	resp.Msg = s.Impl.ColorTransANSI(msg)
	return nil
}

func (s *TerminalModuleRPCServer) SubscribeLines(args *TerminalSubscribeArgs, resp *TerminalSubscribeResp) error {
	if s == nil || s.Impl == nil || s.broker == nil || args == nil || resp == nil {
		return nil
	}
	if args.CallbackBrokerID == 0 {
		return errors.New("TerminalModuleRPCServer.SubscribeLines: callback broker id is 0")
	}

	conn, err := s.broker.Dial(args.CallbackBrokerID)
	if err != nil {
		return err
	}
	cb := &terminalLineCallbackClient{c: rpc.NewClient(conn)}

	ctx, cancel := context.WithCancel(context.Background())
	lines, err := s.Impl.SubscribeLines(ctx)
	if err != nil {
		cancel()
		cb.Stop()
		_ = cb.Close()
		return err
	}

	subID := fmt.Sprintf("sub:%d", atomic.AddUint64(&s.seq, 1))

	s.mu.Lock()
	if s.subs == nil {
		s.subs = make(map[string]context.CancelFunc)
	}
	if s.callbacks == nil {
		s.callbacks = make(map[string]*terminalLineCallbackClient)
	}
	s.subs[subID] = cancel
	s.callbacks[subID] = cb
	s.mu.Unlock()

	go func() {
		defer func() {
			s.mu.Lock()
			if s.subs != nil {
				delete(s.subs, subID)
			}
			if s.callbacks != nil {
				delete(s.callbacks, subID)
			}
			s.mu.Unlock()

			cb.Stop()
			_ = cb.Close()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-lines:
				if !ok {
					return
				}
				_ = cb.OnLine(line)
			}
		}
	}()

	resp.SubID = subID
	return nil
}

func (s *TerminalModuleRPCServer) UnsubscribeLines(args *TerminalUnsubscribeArgs, resp *BoolResp) error {
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
	cb := (*terminalLineCallbackClient)(nil)
	if s.callbacks != nil {
		cb = s.callbacks[args.SubID]
		delete(s.callbacks, args.SubID)
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

func (s *TerminalModuleRPCServer) InterceptNextLine(args *TerminalInterceptArgs, resp *TerminalInterceptResp) error {
	if s == nil || s.Impl == nil || s.broker == nil || args == nil || resp == nil {
		return nil
	}
	if args.CallbackBrokerID == 0 {
		return errors.New("TerminalModuleRPCServer.InterceptNextLine: callback broker id is 0")
	}

	conn, err := s.broker.Dial(args.CallbackBrokerID)
	if err != nil {
		return err
	}
	cb := &terminalLineCallbackClient{c: rpc.NewClient(conn)}

	timeout := time.Duration(0)
	if args.TimeoutMillis > 0 {
		timeout = time.Duration(args.TimeoutMillis) * time.Millisecond
	}

	ctx, cancelCtx := context.WithCancel(context.Background())
	interceptID := fmt.Sprintf("intercept:%d", atomic.AddUint64(&s.interceptSeq, 1))

	cancel, err := s.Impl.InterceptNextLine(ctx, timeout, func(line string) {
		_ = cb.OnLine(line)
		cb.Stop()
		_ = cb.Close()

		s.interceptMu.Lock()
		if s.interceptCancels != nil {
			delete(s.interceptCancels, interceptID)
		}
		if s.interceptCallbacks != nil {
			delete(s.interceptCallbacks, interceptID)
		}
		s.interceptMu.Unlock()
	})
	if err != nil {
		cancelCtx()
		cb.Stop()
		_ = cb.Close()
		return err
	}

	s.interceptMu.Lock()
	if s.interceptCancels == nil {
		s.interceptCancels = make(map[string]func())
	}
	if s.interceptCallbacks == nil {
		s.interceptCallbacks = make(map[string]*terminalLineCallbackClient)
	}
	s.interceptCancels[interceptID] = func() {
		cancelCtx()
		if cancel != nil {
			cancel()
		}
	}
	s.interceptCallbacks[interceptID] = cb
	s.interceptMu.Unlock()

	resp.InterceptID = interceptID
	return nil
}

func (s *TerminalModuleRPCServer) CancelIntercept(args *TerminalCancelInterceptArgs, resp *BoolResp) error {
	if s == nil || args == nil {
		return nil
	}
	if args.InterceptID == "" {
		return nil
	}

	var cancel func()
	var cb *terminalLineCallbackClient
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
		cb.Stop()
		_ = cb.Close()
	}
	if resp != nil {
		resp.OK = ok
	}
	return nil
}

type terminalModuleRPCClient struct {
	c      *rpc.Client
	broker *plugin.MuxBroker
	mu     sync.Mutex
}

func newTerminalModuleRPCClient(conn net.Conn, broker *plugin.MuxBroker) api.TerminalModule {
	if conn == nil {
		return nil
	}
	return &terminalModuleRPCClient{
		c:      rpc.NewClient(conn),
		broker: broker,
	}
}

func (c *terminalModuleRPCClient) Name() string { return api.NameTerminalModule }

func (c *terminalModuleRPCClient) Print(level api.Level, scope string, msg string) {
	if c == nil || c.c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.c.Call("Plugin.Print", &TerminalPrintArgs{Level: level, Scope: scope, Msg: msg}, &Empty{})
}

func (c *terminalModuleRPCClient) Info(scope string, msg string) {
	c.Print(api.LevelInfo, scope, msg)
}
func (c *terminalModuleRPCClient) Warn(scope string, msg string) {
	c.Print(api.LevelWarn, scope, msg)
}
func (c *terminalModuleRPCClient) Error(scope string, msg string) {
	c.Print(api.LevelError, scope, msg)
}
func (c *terminalModuleRPCClient) Success(scope string, msg string) {
	c.Print(api.LevelSuccess, scope, msg)
}

func (c *terminalModuleRPCClient) Raw(msg string) {
	if c == nil || c.c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.c.Call("Plugin.Raw", &TerminalRawArgs{Msg: msg}, &Empty{})
}

func (c *terminalModuleRPCClient) ColorTransANSI(msg string) string {
	if c == nil || c.c == nil {
		return msg
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	resp := TerminalColorTransResp{Msg: msg}
	if err := c.c.Call("Plugin.ColorTransANSI", &TerminalColorTransArgs{Msg: msg}, &resp); err != nil {
		return msg
	}
	return resp.Msg
}

func (c *terminalModuleRPCClient) SubscribeLines(ctx context.Context) (<-chan string, error) {
	if c == nil || c.c == nil || c.broker == nil {
		return nil, errors.New("terminalModuleRPCClient.SubscribeLines: client is not initialised")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	out := make(chan string, 64)
	cbID := c.broker.NextId()
	cbSrv := &terminalLineCallbackServer{ch: out}
	go acceptAndServeMuxBroker(c.broker, cbID, cbSrv)

	c.mu.Lock()
	var resp TerminalSubscribeResp
	err := c.c.Call("Plugin.SubscribeLines", &TerminalSubscribeArgs{CallbackBrokerID: cbID}, &resp)
	c.mu.Unlock()
	if err != nil {
		_ = cbSrv.Stop(&Empty{}, &Empty{})
		return nil, err
	}
	if resp.SubID == "" {
		_ = cbSrv.Stop(&Empty{}, &Empty{})
		return nil, errors.New("terminalModuleRPCClient.SubscribeLines: empty sub id")
	}

	var once sync.Once
	stop := func() {
		once.Do(func() {
			c.mu.Lock()
			_ = c.c.Call("Plugin.UnsubscribeLines", &TerminalUnsubscribeArgs{SubID: resp.SubID}, &BoolResp{})
			c.mu.Unlock()
			_ = cbSrv.Stop(&Empty{}, &Empty{})
		})
	}
	go func() {
		<-ctx.Done()
		stop()
	}()

	return out, nil
}

func (c *terminalModuleRPCClient) InterceptNextLine(ctx context.Context, timeout time.Duration, handler func(string)) (func(), error) {
	if c == nil || c.c == nil || c.broker == nil {
		return nil, errors.New("terminalModuleRPCClient.InterceptNextLine: client is not initialised")
	}
	if handler == nil {
		return nil, errors.New("terminalModuleRPCClient.InterceptNextLine: handler is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	ch := make(chan string, 1)
	cbID := c.broker.NextId()
	cbSrv := &terminalLineCallbackServer{ch: ch}
	go acceptAndServeMuxBroker(c.broker, cbID, cbSrv)

	timeoutMs := int64(0)
	if timeout > 0 {
		timeoutMs = timeout.Milliseconds()
	}

	c.mu.Lock()
	var resp TerminalInterceptResp
	err := c.c.Call("Plugin.InterceptNextLine", &TerminalInterceptArgs{CallbackBrokerID: cbID, TimeoutMillis: timeoutMs}, &resp)
	c.mu.Unlock()
	if err != nil {
		_ = cbSrv.Stop(&Empty{}, &Empty{})
		return nil, err
	}

	go func() {
		select {
		case <-ctx.Done():
			return
		case line, ok := <-ch:
			if !ok {
				return
			}
			handler(line)
		}
	}()

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			if resp.InterceptID != "" {
				c.mu.Lock()
				_ = c.c.Call("Plugin.CancelIntercept", &TerminalCancelInterceptArgs{InterceptID: resp.InterceptID}, &BoolResp{})
				c.mu.Unlock()
			}
			_ = cbSrv.Stop(&Empty{}, &Empty{})
		})
	}
	return cancel, nil
}
