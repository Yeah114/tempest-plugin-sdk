package protocol

import (
	"errors"
	"net"
	"net/rpc"
	"sync"
	"time"

	"github.com/Yeah114/EmptyDea-plugin-sdk/api"
)

type CommandsModuleNameResp struct {
	Name string
}

type CommandsSendArgs struct {
	Command     string
	Dimensional bool
}

type CommandsSendWithRespArgs struct {
	Command      string
	TimeoutNanos int64
}

type CommandsSendWithRespResp struct {
	Output *api.CommandOutput
}

type CommandsTitleArgs struct {
	Message string
}

type CommandsChatArgs struct {
	Content string
}

type CommandsModuleRPCServer struct {
	Impl api.CommandsModule
}

func (s *CommandsModuleRPCServer) Name(_ *Empty, resp *CommandsModuleNameResp) error {
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

func (s *CommandsModuleRPCServer) SendSettingsCommand(args *CommandsSendArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	return s.Impl.SendSettingsCommand(args.Command, args.Dimensional)
}

func (s *CommandsModuleRPCServer) SendPlayerCommand(args *CommandsSendArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	return s.Impl.SendPlayerCommand(args.Command)
}

func (s *CommandsModuleRPCServer) SendWSCommand(args *CommandsSendArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	return s.Impl.SendWSCommand(args.Command)
}

func (s *CommandsModuleRPCServer) SendPlayerCommandWithResp(args *CommandsSendWithRespArgs, resp *CommandsSendWithRespResp) error {
	if resp == nil {
		return nil
	}
	resp.Output = nil
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	out, err := s.Impl.SendPlayerCommandWithResp(args.Command, time.Duration(args.TimeoutNanos))
	if err != nil {
		return err
	}
	resp.Output = out
	return nil
}

func (s *CommandsModuleRPCServer) SendWSCommandWithResp(args *CommandsSendWithRespArgs, resp *CommandsSendWithRespResp) error {
	if resp == nil {
		return nil
	}
	resp.Output = nil
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	out, err := s.Impl.SendWSCommandWithResp(args.Command, time.Duration(args.TimeoutNanos))
	if err != nil {
		return err
	}
	resp.Output = out
	return nil
}

func (s *CommandsModuleRPCServer) AwaitChangesGeneral(_ *Empty, _ *Empty) error {
	if s == nil || s.Impl == nil {
		return nil
	}
	return s.Impl.AwaitChangesGeneral()
}

func (s *CommandsModuleRPCServer) SendChat(args *CommandsChatArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	return s.Impl.SendChat(args.Content)
}

func (s *CommandsModuleRPCServer) Title(args *CommandsTitleArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	return s.Impl.Title(args.Message)
}

type commandsModuleRPCClient struct {
	c  *rpc.Client
	mu sync.Mutex
}

func newCommandsModuleRPCClient(conn net.Conn) api.CommandsModule {
	if conn == nil {
		return nil
	}
	return &commandsModuleRPCClient{c: rpc.NewClient(conn)}
}

func (c *commandsModuleRPCClient) Name() string { return api.NameCommandsModule }

func (c *commandsModuleRPCClient) callNoResp(method string, args any) error {
	if c == nil || c.c == nil {
		return errors.New("commandsModuleRPCClient: client is not initialised")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call(method, args, &Empty{})
}

func (c *commandsModuleRPCClient) SendSettingsCommand(command string, dimensional bool) error {
	return c.callNoResp("Plugin.SendSettingsCommand", &CommandsSendArgs{Command: command, Dimensional: dimensional})
}
func (c *commandsModuleRPCClient) SendPlayerCommand(command string) error {
	return c.callNoResp("Plugin.SendPlayerCommand", &CommandsSendArgs{Command: command})
}
func (c *commandsModuleRPCClient) SendWSCommand(command string) error {
	return c.callNoResp("Plugin.SendWSCommand", &CommandsSendArgs{Command: command})
}

func (c *commandsModuleRPCClient) SendPlayerCommandWithResp(command string, timeout time.Duration) (*api.CommandOutput, error) {
	if c == nil || c.c == nil {
		return nil, errors.New("commandsModuleRPCClient: client is not initialised")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp CommandsSendWithRespResp
	if err := c.c.Call("Plugin.SendPlayerCommandWithResp", &CommandsSendWithRespArgs{Command: command, TimeoutNanos: timeout.Nanoseconds()}, &resp); err != nil {
		return nil, err
	}
	return resp.Output, nil
}

func (c *commandsModuleRPCClient) SendWSCommandWithResp(command string, timeout time.Duration) (*api.CommandOutput, error) {
	if c == nil || c.c == nil {
		return nil, errors.New("commandsModuleRPCClient: client is not initialised")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp CommandsSendWithRespResp
	if err := c.c.Call("Plugin.SendWSCommandWithResp", &CommandsSendWithRespArgs{Command: command, TimeoutNanos: timeout.Nanoseconds()}, &resp); err != nil {
		return nil, err
	}
	return resp.Output, nil
}

func (c *commandsModuleRPCClient) AwaitChangesGeneral() error {
	return c.callNoResp("Plugin.AwaitChangesGeneral", &Empty{})
}

func (c *commandsModuleRPCClient) SendChat(content string) error {
	return c.callNoResp("Plugin.SendChat", &CommandsChatArgs{Content: content})
}

func (c *commandsModuleRPCClient) Title(message string) error {
	return c.callNoResp("Plugin.Title", &CommandsTitleArgs{Message: message})
}
