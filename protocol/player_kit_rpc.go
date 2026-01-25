package protocol

import (
	"context"
	"net"
	"net/rpc"
	"sync"
	"time"

	"github.com/Yeah114/EmptyDea-plugin-sdk/api"
)

type PlayerKitTimeoutArgs struct {
	TimeoutMs int64
}

type PlayerKitSetBoolArgs struct {
	TimeoutMs int64
	Allow     bool
}

type PlayerKitRawSayArgs struct {
	JSONText string
}

type PlayerKitSayArgs struct {
	Message string
}

type PlayerKitSubtitleArgs struct {
	SubtitleMessage string
	TitleMessage    string
}

type PlayerKitStringResp struct {
	Value string
}

type PlayerKitInt64Resp struct {
	Value int64
}

type PlayerKitInt32Resp struct {
	Value int32
}

type PlayerKitUint64Resp struct {
	Value uint64
}

type PlayerKitTimeResp struct {
	Value time.Time
}

type PlayerKitBoolResp struct {
	Value bool
}

type PlayerKitMetadataResp struct {
	Value map[uint32]any
}

type PlayerKitRPCServer struct {
	Impl api.PlayerKit
}

func ctxFromTimeoutMs(timeoutMs int64) (context.Context, func()) {
	if timeoutMs <= 0 {
		return context.Background(), func() {}
	}
	return context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
}

func (s *PlayerKitRPCServer) GetUUIDString(_ *Empty, resp *PlayerKitStringResp) error {
	if resp == nil {
		return nil
	}
	resp.Value = ""
	if s == nil || s.Impl == nil {
		return nil
	}
	resp.Value = s.Impl.GetUUIDString()
	return nil
}

func (s *PlayerKitRPCServer) GetName(_ *Empty, resp *PlayerKitStringResp) error {
	if resp == nil {
		return nil
	}
	resp.Value = ""
	if s == nil || s.Impl == nil {
		return nil
	}
	resp.Value = s.Impl.GetName()
	return nil
}

func (s *PlayerKitRPCServer) GetEntityUniqueID(args *PlayerKitTimeoutArgs, resp *PlayerKitInt64Resp) error {
	if resp == nil {
		return nil
	}
	resp.Value = 0
	if s == nil || s.Impl == nil {
		return nil
	}
	ctx, cancel := ctxFromTimeoutMs(0)
	if args != nil {
		ctx, cancel = ctxFromTimeoutMs(args.TimeoutMs)
	}
	defer cancel()
	v, err := s.Impl.GetEntityUniqueID(ctx)
	if err != nil {
		return err
	}
	resp.Value = v
	return nil
}

func (s *PlayerKitRPCServer) GetLoginTime(args *PlayerKitTimeoutArgs, resp *PlayerKitTimeResp) error {
	if resp == nil {
		return nil
	}
	resp.Value = time.Time{}
	if s == nil || s.Impl == nil {
		return nil
	}
	ctx, cancel := ctxFromTimeoutMs(0)
	if args != nil {
		ctx, cancel = ctxFromTimeoutMs(args.TimeoutMs)
	}
	defer cancel()
	v, err := s.Impl.GetLoginTime(ctx)
	if err != nil {
		return err
	}
	resp.Value = v
	return nil
}

func (s *PlayerKitRPCServer) GetPlatformChatID(args *PlayerKitTimeoutArgs, resp *PlayerKitStringResp) error {
	if resp == nil {
		return nil
	}
	resp.Value = ""
	if s == nil || s.Impl == nil {
		return nil
	}
	ctx, cancel := ctxFromTimeoutMs(0)
	if args != nil {
		ctx, cancel = ctxFromTimeoutMs(args.TimeoutMs)
	}
	defer cancel()
	v, err := s.Impl.GetPlatformChatID(ctx)
	if err != nil {
		return err
	}
	resp.Value = v
	return nil
}

func (s *PlayerKitRPCServer) GetBuildPlatform(args *PlayerKitTimeoutArgs, resp *PlayerKitInt32Resp) error {
	if resp == nil {
		return nil
	}
	resp.Value = 0
	if s == nil || s.Impl == nil {
		return nil
	}
	ctx, cancel := ctxFromTimeoutMs(0)
	if args != nil {
		ctx, cancel = ctxFromTimeoutMs(args.TimeoutMs)
	}
	defer cancel()
	v, err := s.Impl.GetBuildPlatform(ctx)
	if err != nil {
		return err
	}
	resp.Value = v
	return nil
}

func (s *PlayerKitRPCServer) GetSkinID(args *PlayerKitTimeoutArgs, resp *PlayerKitStringResp) error {
	if resp == nil {
		return nil
	}
	resp.Value = ""
	if s == nil || s.Impl == nil {
		return nil
	}
	ctx, cancel := ctxFromTimeoutMs(0)
	if args != nil {
		ctx, cancel = ctxFromTimeoutMs(args.TimeoutMs)
	}
	defer cancel()
	v, err := s.Impl.GetSkinID(ctx)
	if err != nil {
		return err
	}
	resp.Value = v
	return nil
}

func (s *PlayerKitRPCServer) GetCanBuild(args *PlayerKitTimeoutArgs, resp *PlayerKitBoolResp) error {
	return s.getBool(args, resp, func(ctx context.Context) (bool, error) { return s.Impl.GetCanBuild(ctx) })
}

func (s *PlayerKitRPCServer) SetCanBuild(args *PlayerKitSetBoolArgs, _ *Empty) error {
	return s.setBool(args, func(ctx context.Context, allow bool) error { return s.Impl.SetCanBuild(ctx, allow) })
}

func (s *PlayerKitRPCServer) GetCanDig(args *PlayerKitTimeoutArgs, resp *PlayerKitBoolResp) error {
	return s.getBool(args, resp, func(ctx context.Context) (bool, error) { return s.Impl.GetCanDig(ctx) })
}

func (s *PlayerKitRPCServer) SetCanDig(args *PlayerKitSetBoolArgs, _ *Empty) error {
	return s.setBool(args, func(ctx context.Context, allow bool) error { return s.Impl.SetCanDig(ctx, allow) })
}

func (s *PlayerKitRPCServer) GetCanUseDoorsAndSwitches(args *PlayerKitTimeoutArgs, resp *PlayerKitBoolResp) error {
	return s.getBool(args, resp, func(ctx context.Context) (bool, error) { return s.Impl.GetCanUseDoorsAndSwitches(ctx) })
}

func (s *PlayerKitRPCServer) SetCanUseDoorsAndSwitches(args *PlayerKitSetBoolArgs, _ *Empty) error {
	return s.setBool(args, func(ctx context.Context, allow bool) error { return s.Impl.SetCanUseDoorsAndSwitches(ctx, allow) })
}

func (s *PlayerKitRPCServer) GetCanOpenContainers(args *PlayerKitTimeoutArgs, resp *PlayerKitBoolResp) error {
	return s.getBool(args, resp, func(ctx context.Context) (bool, error) { return s.Impl.GetCanOpenContainers(ctx) })
}

func (s *PlayerKitRPCServer) SetCanOpenContainers(args *PlayerKitSetBoolArgs, _ *Empty) error {
	return s.setBool(args, func(ctx context.Context, allow bool) error { return s.Impl.SetCanOpenContainers(ctx, allow) })
}

func (s *PlayerKitRPCServer) GetCanAttackPlayers(args *PlayerKitTimeoutArgs, resp *PlayerKitBoolResp) error {
	return s.getBool(args, resp, func(ctx context.Context) (bool, error) { return s.Impl.GetCanAttackPlayers(ctx) })
}

func (s *PlayerKitRPCServer) SetCanAttackPlayers(args *PlayerKitSetBoolArgs, _ *Empty) error {
	return s.setBool(args, func(ctx context.Context, allow bool) error { return s.Impl.SetCanAttackPlayers(ctx, allow) })
}

func (s *PlayerKitRPCServer) GetCanAttackMobs(args *PlayerKitTimeoutArgs, resp *PlayerKitBoolResp) error {
	return s.getBool(args, resp, func(ctx context.Context) (bool, error) { return s.Impl.GetCanAttackMobs(ctx) })
}

func (s *PlayerKitRPCServer) SetCanAttackMobs(args *PlayerKitSetBoolArgs, _ *Empty) error {
	return s.setBool(args, func(ctx context.Context, allow bool) error { return s.Impl.SetCanAttackMobs(ctx, allow) })
}

func (s *PlayerKitRPCServer) GetCanUseOperatorCommands(args *PlayerKitTimeoutArgs, resp *PlayerKitBoolResp) error {
	return s.getBool(args, resp, func(ctx context.Context) (bool, error) { return s.Impl.GetCanUseOperatorCommands(ctx) })
}

func (s *PlayerKitRPCServer) SetCanUseOperatorCommands(args *PlayerKitSetBoolArgs, _ *Empty) error {
	return s.setBool(args, func(ctx context.Context, allow bool) error { return s.Impl.SetCanUseOperatorCommands(ctx, allow) })
}

func (s *PlayerKitRPCServer) GetCanTeleport(args *PlayerKitTimeoutArgs, resp *PlayerKitBoolResp) error {
	return s.getBool(args, resp, func(ctx context.Context) (bool, error) { return s.Impl.GetCanTeleport(ctx) })
}

type PlayerKitSetCanTeleportResp struct {
	Value bool
}

func (s *PlayerKitRPCServer) SetCanTeleport(args *PlayerKitSetBoolArgs, resp *PlayerKitSetCanTeleportResp) error {
	if resp == nil {
		return nil
	}
	resp.Value = false
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	ctx, cancel := ctxFromTimeoutMs(args.TimeoutMs)
	defer cancel()
	v, err := s.Impl.SetCanTeleport(ctx, args.Allow)
	if err != nil {
		return err
	}
	resp.Value = v
	return nil
}

func (s *PlayerKitRPCServer) GetStatusInvulnerable(args *PlayerKitTimeoutArgs, resp *PlayerKitBoolResp) error {
	return s.getBool(args, resp, func(ctx context.Context) (bool, error) { return s.Impl.GetStatusInvulnerable(ctx) })
}

func (s *PlayerKitRPCServer) GetStatusFlying(args *PlayerKitTimeoutArgs, resp *PlayerKitBoolResp) error {
	return s.getBool(args, resp, func(ctx context.Context) (bool, error) { return s.Impl.GetStatusFlying(ctx) })
}

func (s *PlayerKitRPCServer) GetStatusMayFly(args *PlayerKitTimeoutArgs, resp *PlayerKitBoolResp) error {
	return s.getBool(args, resp, func(ctx context.Context) (bool, error) { return s.Impl.GetStatusMayFly(ctx) })
}

func (s *PlayerKitRPCServer) GetDeviceID(args *PlayerKitTimeoutArgs, resp *PlayerKitStringResp) error {
	if resp == nil {
		return nil
	}
	resp.Value = ""
	if s == nil || s.Impl == nil {
		return nil
	}
	ctx, cancel := ctxFromTimeoutMs(0)
	if args != nil {
		ctx, cancel = ctxFromTimeoutMs(args.TimeoutMs)
	}
	defer cancel()
	v, err := s.Impl.GetDeviceID(ctx)
	if err != nil {
		return err
	}
	resp.Value = v
	return nil
}

func (s *PlayerKitRPCServer) GetEntityRuntimeID(args *PlayerKitTimeoutArgs, resp *PlayerKitUint64Resp) error {
	if resp == nil {
		return nil
	}
	resp.Value = 0
	if s == nil || s.Impl == nil {
		return nil
	}
	ctx, cancel := ctxFromTimeoutMs(0)
	if args != nil {
		ctx, cancel = ctxFromTimeoutMs(args.TimeoutMs)
	}
	defer cancel()
	v, err := s.Impl.GetEntityRuntimeID(ctx)
	if err != nil {
		return err
	}
	resp.Value = v
	return nil
}

func (s *PlayerKitRPCServer) GetEntityMetadata(args *PlayerKitTimeoutArgs, resp *PlayerKitMetadataResp) error {
	if resp == nil {
		return nil
	}
	resp.Value = nil
	if s == nil || s.Impl == nil {
		return nil
	}
	ctx, cancel := ctxFromTimeoutMs(0)
	if args != nil {
		ctx, cancel = ctxFromTimeoutMs(args.TimeoutMs)
	}
	defer cancel()
	v, err := s.Impl.GetEntityMetadata(ctx)
	if err != nil {
		return err
	}
	resp.Value = v
	return nil
}

func (s *PlayerKitRPCServer) GetIsOP(args *PlayerKitTimeoutArgs, resp *PlayerKitBoolResp) error {
	return s.getBool(args, resp, func(ctx context.Context) (bool, error) { return s.Impl.GetIsOP(ctx) })
}

func (s *PlayerKitRPCServer) GetOnline(args *PlayerKitTimeoutArgs, resp *PlayerKitBoolResp) error {
	return s.getBool(args, resp, func(ctx context.Context) (bool, error) { return s.Impl.GetOnline(ctx) })
}

func (s *PlayerKitRPCServer) RawSay(args *PlayerKitRawSayArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	return s.Impl.RawSay(args.JSONText)
}

func (s *PlayerKitRPCServer) Say(args *PlayerKitSayArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	return s.Impl.Say(args.Message)
}

func (s *PlayerKitRPCServer) Title(args *PlayerKitSayArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	return s.Impl.Title(args.Message)
}

func (s *PlayerKitRPCServer) Subtitle(args *PlayerKitSubtitleArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	return s.Impl.Subtitle(args.SubtitleMessage, args.TitleMessage)
}

func (s *PlayerKitRPCServer) ActionBar(args *PlayerKitSayArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	return s.Impl.ActionBar(args.Message)
}

func (s *PlayerKitRPCServer) getBool(args *PlayerKitTimeoutArgs, resp *PlayerKitBoolResp, fn func(context.Context) (bool, error)) error {
	if resp == nil {
		return nil
	}
	resp.Value = false
	if s == nil || s.Impl == nil || fn == nil {
		return nil
	}
	ctx, cancel := ctxFromTimeoutMs(0)
	if args != nil {
		ctx, cancel = ctxFromTimeoutMs(args.TimeoutMs)
	}
	defer cancel()
	v, err := fn(ctx)
	if err != nil {
		return err
	}
	resp.Value = v
	return nil
}

func (s *PlayerKitRPCServer) setBool(args *PlayerKitSetBoolArgs, fn func(context.Context, bool) error) error {
	if s == nil || s.Impl == nil || fn == nil || args == nil {
		return nil
	}
	ctx, cancel := ctxFromTimeoutMs(args.TimeoutMs)
	defer cancel()
	return fn(ctx, args.Allow)
}

type playerKitRPCClient struct {
	c  *rpc.Client
	mu sync.Mutex
}

func newPlayerKitRPCClient(conn net.Conn) api.PlayerKit {
	if conn == nil {
		return nil
	}
	return &playerKitRPCClient{c: rpc.NewClient(conn)}
}

func timeoutMsFromContext(ctx context.Context) int64 {
	if ctx == nil {
		return 0
	}
	if deadline, ok := ctx.Deadline(); ok {
		d := time.Until(deadline)
		if d <= 0 {
			return 1
		}
		return d.Milliseconds()
	}
	return 0
}

func (c *playerKitRPCClient) GetUUIDString() string {
	if c == nil || c.c == nil {
		return ""
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp PlayerKitStringResp
	_ = c.c.Call("Plugin.GetUUIDString", &Empty{}, &resp)
	return resp.Value
}

func (c *playerKitRPCClient) GetName() string {
	if c == nil || c.c == nil {
		return ""
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp PlayerKitStringResp
	_ = c.c.Call("Plugin.GetName", &Empty{}, &resp)
	return resp.Value
}

func (c *playerKitRPCClient) GetEntityUniqueID(ctx context.Context) (int64, error) {
	var resp PlayerKitInt64Resp
	err := c.callWithTimeout("Plugin.GetEntityUniqueID", ctx, &PlayerKitTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}, &resp)
	return resp.Value, err
}

func (c *playerKitRPCClient) GetLoginTime(ctx context.Context) (time.Time, error) {
	var resp PlayerKitTimeResp
	err := c.callWithTimeout("Plugin.GetLoginTime", ctx, &PlayerKitTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}, &resp)
	return resp.Value, err
}

func (c *playerKitRPCClient) GetPlatformChatID(ctx context.Context) (string, error) {
	var resp PlayerKitStringResp
	err := c.callWithTimeout("Plugin.GetPlatformChatID", ctx, &PlayerKitTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}, &resp)
	return resp.Value, err
}

func (c *playerKitRPCClient) GetBuildPlatform(ctx context.Context) (int32, error) {
	var resp PlayerKitInt32Resp
	err := c.callWithTimeout("Plugin.GetBuildPlatform", ctx, &PlayerKitTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}, &resp)
	return resp.Value, err
}

func (c *playerKitRPCClient) GetSkinID(ctx context.Context) (string, error) {
	var resp PlayerKitStringResp
	err := c.callWithTimeout("Plugin.GetSkinID", ctx, &PlayerKitTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}, &resp)
	return resp.Value, err
}

func (c *playerKitRPCClient) GetCanBuild(ctx context.Context) (bool, error) {
	return c.getBool("Plugin.GetCanBuild", ctx)
}
func (c *playerKitRPCClient) SetCanBuild(ctx context.Context, allow bool) error {
	return c.setBool("Plugin.SetCanBuild", ctx, allow)
}
func (c *playerKitRPCClient) GetCanDig(ctx context.Context) (bool, error) {
	return c.getBool("Plugin.GetCanDig", ctx)
}
func (c *playerKitRPCClient) SetCanDig(ctx context.Context, allow bool) error {
	return c.setBool("Plugin.SetCanDig", ctx, allow)
}
func (c *playerKitRPCClient) GetCanUseDoorsAndSwitches(ctx context.Context) (bool, error) {
	return c.getBool("Plugin.GetCanUseDoorsAndSwitches", ctx)
}
func (c *playerKitRPCClient) SetCanUseDoorsAndSwitches(ctx context.Context, allow bool) error {
	return c.setBool("Plugin.SetCanUseDoorsAndSwitches", ctx, allow)
}
func (c *playerKitRPCClient) GetCanOpenContainers(ctx context.Context) (bool, error) {
	return c.getBool("Plugin.GetCanOpenContainers", ctx)
}
func (c *playerKitRPCClient) SetCanOpenContainers(ctx context.Context, allow bool) error {
	return c.setBool("Plugin.SetCanOpenContainers", ctx, allow)
}
func (c *playerKitRPCClient) GetCanAttackPlayers(ctx context.Context) (bool, error) {
	return c.getBool("Plugin.GetCanAttackPlayers", ctx)
}
func (c *playerKitRPCClient) SetCanAttackPlayers(ctx context.Context, allow bool) error {
	return c.setBool("Plugin.SetCanAttackPlayers", ctx, allow)
}
func (c *playerKitRPCClient) GetCanAttackMobs(ctx context.Context) (bool, error) {
	return c.getBool("Plugin.GetCanAttackMobs", ctx)
}
func (c *playerKitRPCClient) SetCanAttackMobs(ctx context.Context, allow bool) error {
	return c.setBool("Plugin.SetCanAttackMobs", ctx, allow)
}
func (c *playerKitRPCClient) GetCanUseOperatorCommands(ctx context.Context) (bool, error) {
	return c.getBool("Plugin.GetCanUseOperatorCommands", ctx)
}
func (c *playerKitRPCClient) SetCanUseOperatorCommands(ctx context.Context, allow bool) error {
	return c.setBool("Plugin.SetCanUseOperatorCommands", ctx, allow)
}
func (c *playerKitRPCClient) GetCanTeleport(ctx context.Context) (bool, error) {
	return c.getBool("Plugin.GetCanTeleport", ctx)
}

func (c *playerKitRPCClient) SetCanTeleport(ctx context.Context, allow bool) (bool, error) {
	if c == nil || c.c == nil {
		return false, nil
	}
	args := &PlayerKitSetBoolArgs{TimeoutMs: timeoutMsFromContext(ctx), Allow: allow}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp PlayerKitSetCanTeleportResp
	err := c.c.Call("Plugin.SetCanTeleport", args, &resp)
	return resp.Value, err
}

func (c *playerKitRPCClient) GetStatusInvulnerable(ctx context.Context) (bool, error) {
	return c.getBool("Plugin.GetStatusInvulnerable", ctx)
}
func (c *playerKitRPCClient) GetStatusFlying(ctx context.Context) (bool, error) {
	return c.getBool("Plugin.GetStatusFlying", ctx)
}
func (c *playerKitRPCClient) GetStatusMayFly(ctx context.Context) (bool, error) {
	return c.getBool("Plugin.GetStatusMayFly", ctx)
}

func (c *playerKitRPCClient) GetDeviceID(ctx context.Context) (string, error) {
	if c == nil || c.c == nil {
		return "", nil
	}
	args := &PlayerKitTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp PlayerKitStringResp
	err := c.c.Call("Plugin.GetDeviceID", args, &resp)
	return resp.Value, err
}

func (c *playerKitRPCClient) GetEntityRuntimeID(ctx context.Context) (uint64, error) {
	if c == nil || c.c == nil {
		return 0, nil
	}
	args := &PlayerKitTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp PlayerKitUint64Resp
	err := c.c.Call("Plugin.GetEntityRuntimeID", args, &resp)
	return resp.Value, err
}

func (c *playerKitRPCClient) GetEntityMetadata(ctx context.Context) (map[uint32]any, error) {
	if c == nil || c.c == nil {
		return nil, nil
	}
	args := &PlayerKitTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp PlayerKitMetadataResp
	err := c.c.Call("Plugin.GetEntityMetadata", args, &resp)
	return resp.Value, err
}

func (c *playerKitRPCClient) GetIsOP(ctx context.Context) (bool, error) {
	return c.getBool("Plugin.GetIsOP", ctx)
}
func (c *playerKitRPCClient) GetOnline(ctx context.Context) (bool, error) {
	return c.getBool("Plugin.GetOnline", ctx)
}

func (c *playerKitRPCClient) RawSay(jsonText string) error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.RawSay", &PlayerKitRawSayArgs{JSONText: jsonText}, &Empty{})
}

func (c *playerKitRPCClient) Say(message string) error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.Say", &PlayerKitSayArgs{Message: message}, &Empty{})
}

func (c *playerKitRPCClient) Title(message string) error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.Title", &PlayerKitSayArgs{Message: message}, &Empty{})
}

func (c *playerKitRPCClient) Subtitle(subtitleMessage, titleMessage string) error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.Subtitle", &PlayerKitSubtitleArgs{SubtitleMessage: subtitleMessage, TitleMessage: titleMessage}, &Empty{})
}

func (c *playerKitRPCClient) ActionBar(message string) error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.ActionBar", &PlayerKitSayArgs{Message: message}, &Empty{})
}

func (c *playerKitRPCClient) getBool(method string, ctx context.Context) (bool, error) {
	if c == nil || c.c == nil {
		return false, nil
	}
	args := &PlayerKitTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp PlayerKitBoolResp
	err := c.c.Call(method, args, &resp)
	return resp.Value, err
}

func (c *playerKitRPCClient) setBool(method string, ctx context.Context, allow bool) error {
	if c == nil || c.c == nil {
		return nil
	}
	args := &PlayerKitSetBoolArgs{TimeoutMs: timeoutMsFromContext(ctx), Allow: allow}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call(method, args, &Empty{})
}

func (c *playerKitRPCClient) callWithTimeout(method string, ctx context.Context, args any, resp any) error {
	if c == nil || c.c == nil {
		return nil
	}
	_ = ctx
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call(method, args, resp)
}

var _ api.PlayerKit = (*playerKitRPCClient)(nil)
