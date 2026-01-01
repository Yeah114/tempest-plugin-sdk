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

type FlexModuleNameResp struct {
	Name string
}

type FlexSetArgs struct {
	Key string
	Val string
}

type FlexGetArgs struct {
	Key string
}

type FlexGetResp struct {
	Val string
	OK  bool
}

type FlexPublishArgs struct {
	Topic       string
	PayloadJSON []byte
}

type FlexSubscribeArgs struct {
	Topic           string
	CallbackBrokerID uint32
}

type FlexSubscribeResp struct {
	SubID string
}

type FlexUnsubscribeArgs struct {
	SubID string
}

type FlexTopicEvent struct {
	PayloadJSON []byte
}

type flexTopicCallbackServer struct {
	handler func([]byte)
}

func (s *flexTopicCallbackServer) OnEvent(args *FlexTopicEvent, _ *Empty) error {
	if s == nil || s.handler == nil || args == nil {
		return nil
	}
	s.handler(append([]byte(nil), args.PayloadJSON...))
	return nil
}

type flexTopicCallbackClient struct {
	c  *rpc.Client
	mu sync.Mutex
}

func (c *flexTopicCallbackClient) Close() error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Close()
}

func (c *flexTopicCallbackClient) OnEvent(payload []byte) error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.OnEvent", &FlexTopicEvent{PayloadJSON: append([]byte(nil), payload...)}, &Empty{})
}

type FlexExposeArgs struct {
	APIName          string
	HandlerBrokerID  uint32
}

type FlexExposeResp struct {
	OK bool
}

type FlexUnexposeArgs struct {
	APIName string
}

type FlexUnexposeResp struct {
	OK bool
}

type FlexHandleArgs struct {
	TimeoutMs int64
	ArgsJSON  []byte
}

type FlexHandleResp struct {
	ResultJSON []byte
	ErrStr     string
}

type flexHandlerCallbackServer struct {
	handler func(context.Context, []byte) ([]byte, string)
}

func (s *flexHandlerCallbackServer) Handle(args *FlexHandleArgs, resp *FlexHandleResp) error {
	if resp == nil {
		return nil
	}
	resp.ResultJSON = nil
	resp.ErrStr = ""
	if s == nil || s.handler == nil || args == nil {
		return nil
	}
	ctx := context.Background()
	if args.TimeoutMs > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(args.TimeoutMs)*time.Millisecond)
		defer cancel()
	}
	res, errStr := s.handler(ctx, append([]byte(nil), args.ArgsJSON...))
	if res == nil {
		res = []byte("null")
	}
	resp.ResultJSON = res
	resp.ErrStr = errStr
	return nil
}

type flexHandlerCallbackClient struct {
	c  *rpc.Client
	mu sync.Mutex
}

func (c *flexHandlerCallbackClient) Close() error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Close()
}

func (c *flexHandlerCallbackClient) Handle(timeoutMs int64, argsJSON []byte) ([]byte, string, error) {
	if c == nil || c.c == nil {
		return nil, "", errors.New("flex handler client unavailable")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp FlexHandleResp
	if err := c.c.Call("Plugin.Handle", &FlexHandleArgs{TimeoutMs: timeoutMs, ArgsJSON: append([]byte(nil), argsJSON...)}, &resp); err != nil {
		return nil, "", err
	}
	return resp.ResultJSON, resp.ErrStr, nil
}

type FlexCallArgs struct {
	APIName    string
	TimeoutMs  int64
	ArgsJSON   []byte
}

type FlexCallResp struct {
	ResultJSON []byte
	ErrStr     string
}

type FlexModuleRPCServer struct {
	Impl   api.FlexModule
	broker *plugin.MuxBroker

	mu          sync.Mutex
	subCancels  map[string]func()
	unexposeFns map[string]func()
	subSeq      atomic.Uint64
}

func (s *FlexModuleRPCServer) Name(_ *Empty, resp *FlexModuleNameResp) error {
	if resp == nil {
		return nil
	}
	resp.Name = api.NameFlexModule
	if s == nil || s.Impl == nil {
		return nil
	}
	resp.Name = s.Impl.Name()
	return nil
}

func (s *FlexModuleRPCServer) Set(args *FlexSetArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	s.Impl.Set(args.Key, args.Val)
	return nil
}

func (s *FlexModuleRPCServer) Get(args *FlexGetArgs, resp *FlexGetResp) error {
	if resp == nil {
		return nil
	}
	resp.Val, resp.OK = "", false
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	val, ok := s.Impl.Get(args.Key)
	resp.Val, resp.OK = val, ok
	return nil
}

func (s *FlexModuleRPCServer) Publish(args *FlexPublishArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	s.Impl.Publish(args.Topic, args.PayloadJSON)
	return nil
}

func (s *FlexModuleRPCServer) Subscribe(args *FlexSubscribeArgs, resp *FlexSubscribeResp) error {
	if resp == nil {
		return nil
	}
	resp.SubID = ""
	if s == nil || s.Impl == nil || s.broker == nil || args == nil {
		return nil
	}
	if args.CallbackBrokerID == 0 {
		return errors.New("FlexModuleRPCServer.Subscribe: callback broker id is 0")
	}

	conn, err := s.broker.Dial(args.CallbackBrokerID)
	if err != nil {
		return err
	}
	cb := &flexTopicCallbackClient{c: rpc.NewClient(conn)}

	ctx, cancel := context.WithCancel(context.Background())
	ch := s.Impl.Subscribe(ctx, args.Topic)

	subID := "sub:" + fmt.Sprint(s.subSeq.Add(1))
	s.mu.Lock()
	if s.subCancels == nil {
		s.subCancels = make(map[string]func())
	}
	s.subCancels[subID] = cancel
	s.mu.Unlock()

	go func() {
		defer func() {
			_ = cb.Close()
			cancel()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case payload, ok := <-ch:
				if !ok {
					return
				}
				_ = cb.OnEvent(payload)
			}
		}
	}()

	resp.SubID = subID
	return nil
}

func (s *FlexModuleRPCServer) Unsubscribe(args *FlexUnsubscribeArgs, resp *BoolResp) error {
	if resp == nil {
		return nil
	}
	resp.OK = false
	if s == nil || args == nil || args.SubID == "" {
		return nil
	}
	s.mu.Lock()
	cancel := s.subCancels[args.SubID]
	delete(s.subCancels, args.SubID)
	s.mu.Unlock()
	if cancel != nil {
		cancel()
		resp.OK = true
	}
	return nil
}

func (s *FlexModuleRPCServer) Expose(args *FlexExposeArgs, resp *FlexExposeResp) error {
	if resp == nil {
		return nil
	}
	resp.OK = false
	if s == nil || s.Impl == nil || s.broker == nil || args == nil {
		return nil
	}
	if args.HandlerBrokerID == 0 {
		return errors.New("FlexModuleRPCServer.Expose: handler broker id is 0")
	}
	apiName := args.APIName
	if apiName == "" {
		return errors.New("FlexModuleRPCServer.Expose: api name is empty")
	}

	conn, err := s.broker.Dial(args.HandlerBrokerID)
	if err != nil {
		return err
	}
	cb := &flexHandlerCallbackClient{c: rpc.NewClient(conn)}

	unexpose, err := s.Impl.Expose(apiName, func(ctx context.Context, payload []byte) ([]byte, string) {
		timeoutMs := timeoutMsFromContext(ctx)
		res, errStr, callErr := cb.Handle(timeoutMs, payload)
		if callErr != nil {
			return []byte("null"), callErr.Error()
		}
		return res, errStr
	})
	if err != nil {
		_ = cb.Close()
		return err
	}

	s.mu.Lock()
	if s.unexposeFns == nil {
		s.unexposeFns = make(map[string]func())
	}
	s.unexposeFns[apiName] = func() {
		unexpose()
		_ = cb.Close()
	}
	s.mu.Unlock()

	resp.OK = true
	return nil
}

func (s *FlexModuleRPCServer) Unexpose(args *FlexUnexposeArgs, resp *FlexUnexposeResp) error {
	if resp == nil {
		return nil
	}
	resp.OK = false
	if s == nil || args == nil || args.APIName == "" {
		return nil
	}
	s.mu.Lock()
	fn := s.unexposeFns[args.APIName]
	delete(s.unexposeFns, args.APIName)
	s.mu.Unlock()
	if fn != nil {
		fn()
		resp.OK = true
	}
	return nil
}

func (s *FlexModuleRPCServer) Call(args *FlexCallArgs, resp *FlexCallResp) error {
	if resp == nil {
		return nil
	}
	resp.ResultJSON = nil
	resp.ErrStr = ""
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	ctx, cancel := ctxFromTimeoutMs(args.TimeoutMs)
	defer cancel()
	result, errStr, err := s.Impl.Call(ctx, args.APIName, args.ArgsJSON)
	if err != nil {
		return err
	}
	resp.ResultJSON = result
	resp.ErrStr = errStr
	return nil
}

type flexModuleRPCClient struct {
	c      *rpc.Client
	broker *plugin.MuxBroker
	mu     sync.Mutex
}

func newFlexModuleRPCClient(conn net.Conn, broker *plugin.MuxBroker) api.FlexModule {
	if conn == nil {
		return nil
	}
	return &flexModuleRPCClient{c: rpc.NewClient(conn), broker: broker}
}

func (c *flexModuleRPCClient) Name() string { return api.NameFlexModule }

func (c *flexModuleRPCClient) Set(key string, val string) {
	if c == nil || c.c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.c.Call("Plugin.Set", &FlexSetArgs{Key: key, Val: val}, &Empty{})
}

func (c *flexModuleRPCClient) Get(key string) (string, bool) {
	if c == nil || c.c == nil {
		return "", false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp FlexGetResp
	if err := c.c.Call("Plugin.Get", &FlexGetArgs{Key: key}, &resp); err != nil {
		return "", false
	}
	return resp.Val, resp.OK
}

func (c *flexModuleRPCClient) Publish(topic string, payloadJSON []byte) {
	if c == nil || c.c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.c.Call("Plugin.Publish", &FlexPublishArgs{Topic: topic, PayloadJSON: append([]byte(nil), payloadJSON...)}, &Empty{})
}

func (c *flexModuleRPCClient) Subscribe(ctx context.Context, topic string) <-chan []byte {
	out := make(chan []byte, 256)
	if c == nil || c.c == nil || c.broker == nil {
		close(out)
		return out
	}
	if ctx == nil {
		ctx = context.Background()
	}

	cbID := c.broker.NextId()
	go acceptAndServeMuxBroker(c.broker, cbID, &flexTopicCallbackServer{handler: func(payload []byte) {
		select {
		case out <- payload:
		default:
		}
	}})

	c.mu.Lock()
	var resp FlexSubscribeResp
	err := c.c.Call("Plugin.Subscribe", &FlexSubscribeArgs{Topic: topic, CallbackBrokerID: cbID}, &resp)
	c.mu.Unlock()
	if err != nil || resp.SubID == "" {
		close(out)
		return out
	}

	context.AfterFunc(ctx, func() {
		c.mu.Lock()
		var b BoolResp
		_ = c.c.Call("Plugin.Unsubscribe", &FlexUnsubscribeArgs{SubID: resp.SubID}, &b)
		c.mu.Unlock()
		close(out)
	})

	return out
}

func (c *flexModuleRPCClient) Expose(apiName string, handler func(context.Context, []byte) ([]byte, string)) (func(), error) {
	if c == nil || c.c == nil || c.broker == nil {
		return func() {}, errors.New("flexModuleRPCClient.Expose: client is not initialised")
	}
	if handler == nil {
		return func() {}, errors.New("flexModuleRPCClient.Expose: handler is nil")
	}

	cbID := c.broker.NextId()
	go acceptAndServeMuxBroker(c.broker, cbID, &flexHandlerCallbackServer{handler: handler})

	c.mu.Lock()
	var resp FlexExposeResp
	err := c.c.Call("Plugin.Expose", &FlexExposeArgs{APIName: apiName, HandlerBrokerID: cbID}, &resp)
	c.mu.Unlock()
	if err != nil {
		return func() {}, err
	}
	if !resp.OK {
		return func() {}, errors.New("flexModuleRPCClient.Expose: expose failed")
	}

	return func() {
		c.mu.Lock()
		var unresp FlexUnexposeResp
		_ = c.c.Call("Plugin.Unexpose", &FlexUnexposeArgs{APIName: apiName}, &unresp)
		c.mu.Unlock()
	}, nil
}

func (c *flexModuleRPCClient) Call(ctx context.Context, apiName string, argsJSON []byte) ([]byte, string, error) {
	if c == nil || c.c == nil {
		return nil, "", errors.New("flexModuleRPCClient.Call: client is not initialised")
	}
	timeoutMs := timeoutMsFromContext(ctx)
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp FlexCallResp
	if err := c.c.Call("Plugin.Call", &FlexCallArgs{APIName: apiName, TimeoutMs: timeoutMs, ArgsJSON: append([]byte(nil), argsJSON...)}, &resp); err != nil {
		return nil, "", err
	}
	return resp.ResultJSON, resp.ErrStr, nil
}

var _ api.FlexModule = (*flexModuleRPCClient)(nil)
