package protocol

import (
	"net"
	"net/rpc"
	"sync"

	"github.com/Yeah114/tempest-plugin-sdk/api"
)

type LoggerModuleNameResp struct {
	Name string
}

type LoggerLogArgs struct {
	Scope string
	Level api.Level
	Msg   string
}

type LoggerScopeMsgArgs struct {
	Scope string
	Msg   string
}

type LoggerModuleRPCServer struct {
	Impl api.LoggerModule
}

func (s *LoggerModuleRPCServer) Name(_ *Empty, resp *LoggerModuleNameResp) error {
	if resp == nil {
		return nil
	}
	resp.Name = api.NameLoggerModule
	if s == nil || s.Impl == nil {
		return nil
	}
	resp.Name = s.Impl.Name()
	return nil
}

func (s *LoggerModuleRPCServer) Log(args *LoggerLogArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	s.Impl.Log(args.Scope, args.Level, args.Msg)
	return nil
}

func (s *LoggerModuleRPCServer) Info(args *LoggerScopeMsgArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	s.Impl.Info(args.Scope, args.Msg)
	return nil
}

func (s *LoggerModuleRPCServer) Warn(args *LoggerScopeMsgArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	s.Impl.Warn(args.Scope, args.Msg)
	return nil
}

func (s *LoggerModuleRPCServer) Error(args *LoggerScopeMsgArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	s.Impl.Error(args.Scope, args.Msg)
	return nil
}

func (s *LoggerModuleRPCServer) Success(args *LoggerScopeMsgArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	s.Impl.Success(args.Scope, args.Msg)
	return nil
}

type loggerModuleRPCClient struct {
	c  *rpc.Client
	mu sync.Mutex
}

func newLoggerModuleRPCClient(conn net.Conn) api.LoggerModule {
	if conn == nil {
		return nil
	}
	return &loggerModuleRPCClient{c: rpc.NewClient(conn)}
}

func (c *loggerModuleRPCClient) Name() string { return api.NameLoggerModule }

func (c *loggerModuleRPCClient) Log(scope string, level api.Level, msg string) {
	if c == nil || c.c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.c.Call("Plugin.Log", &LoggerLogArgs{Scope: scope, Level: level, Msg: msg}, &Empty{})
}

func (c *loggerModuleRPCClient) Info(scope, msg string) {
	if c == nil || c.c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.c.Call("Plugin.Info", &LoggerScopeMsgArgs{Scope: scope, Msg: msg}, &Empty{})
}

func (c *loggerModuleRPCClient) Warn(scope, msg string) {
	if c == nil || c.c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.c.Call("Plugin.Warn", &LoggerScopeMsgArgs{Scope: scope, Msg: msg}, &Empty{})
}

func (c *loggerModuleRPCClient) Error(scope, msg string) {
	if c == nil || c.c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.c.Call("Plugin.Error", &LoggerScopeMsgArgs{Scope: scope, Msg: msg}, &Empty{})
}

func (c *loggerModuleRPCClient) Success(scope, msg string) {
	if c == nil || c.c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.c.Call("Plugin.Success", &LoggerScopeMsgArgs{Scope: scope, Msg: msg}, &Empty{})
}

var _ api.LoggerModule = (*loggerModuleRPCClient)(nil)
