package protocol

import (
	"context"
	"errors"
	"net"
	"net/rpc"
	"sync"

	"github.com/Yeah114/EmptyDea-plugin-sdk/api"
)

type UQHolderModuleNameResp struct {
	Name string
}

type UQHolderTimeoutArgs struct {
	TimeoutMs int64
}

type UQHolderStringResp struct {
	Value string
}

type UQHolderUint64Resp struct {
	Value uint64
}

type UQHolderInt64Resp struct {
	Value int64
}

type UQHolderUint16OKResp struct {
	Value uint16
	OK    bool
}

type UQHolderInt32OKResp struct {
	Value int32
	OK    bool
}

type UQHolderUint32OKResp struct {
	Value uint32
	OK    bool
}

type UQHolderInt64OKResp struct {
	Value int64
	OK    bool
}

type UQHolderFloat32OKResp struct {
	Value float32
	OK    bool
}

type UQHolderPositionOKResp struct {
	Value [3]float32
	OK    bool
}

type UQHolderByteOKResp struct {
	Value uint8
	OK    bool
}

type UQHolderMapOKResp struct {
	Value map[string]any
	OK    bool
}

type UQHolderMapResp struct {
	Value map[string]any
}

type UQHolderGameRulesResp struct {
	Value map[string]api.GameRule
}

type UQHolderModuleRPCServer struct {
	Impl api.UQHolderModule
}

func (s *UQHolderModuleRPCServer) Name(_ *Empty, resp *UQHolderModuleNameResp) error {
	if resp == nil {
		return nil
	}
	resp.Name = api.NameUQHolderModule
	if s == nil || s.Impl == nil {
		return nil
	}
	resp.Name = s.Impl.Name()
	return nil
}

func (s *UQHolderModuleRPCServer) BotName(args *UQHolderTimeoutArgs, resp *UQHolderStringResp) error {
	return s.callString(args, resp, func(ctx context.Context) (string, error) { return s.Impl.BotName(ctx) })
}

func (s *UQHolderModuleRPCServer) BotRuntimeID(args *UQHolderTimeoutArgs, resp *UQHolderUint64Resp) error {
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
	v, err := s.Impl.BotRuntimeID(ctx)
	if err != nil {
		return err
	}
	resp.Value = v
	return nil
}

func (s *UQHolderModuleRPCServer) BotUniqueID(args *UQHolderTimeoutArgs, resp *UQHolderInt64Resp) error {
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
	v, err := s.Impl.BotUniqueID(ctx)
	if err != nil {
		return err
	}
	resp.Value = v
	return nil
}

func (s *UQHolderModuleRPCServer) BotIdentity(args *UQHolderTimeoutArgs, resp *UQHolderStringResp) error {
	return s.callString(args, resp, func(ctx context.Context) (string, error) { return s.Impl.BotIdentity(ctx) })
}

func (s *UQHolderModuleRPCServer) BotUUIDStr(args *UQHolderTimeoutArgs, resp *UQHolderStringResp) error {
	return s.callString(args, resp, func(ctx context.Context) (string, error) { return s.Impl.BotUUIDStr(ctx) })
}

func (s *UQHolderModuleRPCServer) BotXUID(args *UQHolderTimeoutArgs, resp *UQHolderStringResp) error {
	return s.callString(args, resp, func(ctx context.Context) (string, error) { return s.Impl.BotXUID(ctx) })
}

func (s *UQHolderModuleRPCServer) BasicRaw(args *UQHolderTimeoutArgs, resp *UQHolderMapResp) error {
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
	v, err := s.Impl.BasicRaw(ctx)
	if err != nil {
		return err
	}
	resp.Value = v
	return nil
}

func (s *UQHolderModuleRPCServer) CompressThreshold(args *UQHolderTimeoutArgs, resp *UQHolderUint16OKResp) error {
	if resp == nil {
		return nil
	}
	resp.Value, resp.OK = 0, false
	if s == nil || s.Impl == nil {
		return nil
	}
	ctx, cancel := ctxFromTimeoutMs(0)
	if args != nil {
		ctx, cancel = ctxFromTimeoutMs(args.TimeoutMs)
	}
	defer cancel()
	v, ok, err := s.Impl.CompressThreshold(ctx)
	if err != nil {
		return err
	}
	resp.Value, resp.OK = v, ok
	return nil
}

func (s *UQHolderModuleRPCServer) WorldGameMode(args *UQHolderTimeoutArgs, resp *UQHolderInt32OKResp) error {
	return s.callInt32OK(args, resp, func(ctx context.Context) (int32, bool, error) { return s.Impl.WorldGameMode(ctx) })
}

func (s *UQHolderModuleRPCServer) GameMode(args *UQHolderTimeoutArgs, resp *UQHolderInt32OKResp) error {
	return s.callInt32OK(args, resp, func(ctx context.Context) (int32, bool, error) { return s.Impl.GameMode(ctx) })
}

func (s *UQHolderModuleRPCServer) WorldDifficulty(args *UQHolderTimeoutArgs, resp *UQHolderUint32OKResp) error {
	if resp == nil {
		return nil
	}
	resp.Value, resp.OK = 0, false
	if s == nil || s.Impl == nil {
		return nil
	}
	ctx, cancel := ctxFromTimeoutMs(0)
	if args != nil {
		ctx, cancel = ctxFromTimeoutMs(args.TimeoutMs)
	}
	defer cancel()
	v, ok, err := s.Impl.WorldDifficulty(ctx)
	if err != nil {
		return err
	}
	resp.Value, resp.OK = v, ok
	return nil
}

func (s *UQHolderModuleRPCServer) Time(args *UQHolderTimeoutArgs, resp *UQHolderInt32OKResp) error {
	return s.callInt32OK(args, resp, func(ctx context.Context) (int32, bool, error) { return s.Impl.Time(ctx) })
}

func (s *UQHolderModuleRPCServer) DayTime(args *UQHolderTimeoutArgs, resp *UQHolderInt32OKResp) error {
	return s.callInt32OK(args, resp, func(ctx context.Context) (int32, bool, error) { return s.Impl.DayTime(ctx) })
}

func (s *UQHolderModuleRPCServer) DayTimePercent(args *UQHolderTimeoutArgs, resp *UQHolderFloat32OKResp) error {
	return s.callFloat32OK(args, resp, func(ctx context.Context) (float32, bool, error) { return s.Impl.DayTimePercent(ctx) })
}

func (s *UQHolderModuleRPCServer) CurrentTick(args *UQHolderTimeoutArgs, resp *UQHolderInt64OKResp) error {
	if resp == nil {
		return nil
	}
	resp.Value, resp.OK = 0, false
	if s == nil || s.Impl == nil {
		return nil
	}
	ctx, cancel := ctxFromTimeoutMs(0)
	if args != nil {
		ctx, cancel = ctxFromTimeoutMs(args.TimeoutMs)
	}
	defer cancel()
	v, ok, err := s.Impl.CurrentTick(ctx)
	if err != nil {
		return err
	}
	resp.Value, resp.OK = v, ok
	return nil
}

func (s *UQHolderModuleRPCServer) SyncRatio(args *UQHolderTimeoutArgs, resp *UQHolderFloat32OKResp) error {
	return s.callFloat32OK(args, resp, func(ctx context.Context) (float32, bool, error) { return s.Impl.SyncRatio(ctx) })
}

func (s *UQHolderModuleRPCServer) BotDimension(args *UQHolderTimeoutArgs, resp *UQHolderInt32OKResp) error {
	return s.callInt32OK(args, resp, func(ctx context.Context) (int32, bool, error) { return s.Impl.BotDimension(ctx) })
}

func (s *UQHolderModuleRPCServer) BotPosition(args *UQHolderTimeoutArgs, resp *UQHolderPositionOKResp) error {
	if resp == nil {
		return nil
	}
	resp.Value, resp.OK = [3]float32{}, false
	if s == nil || s.Impl == nil {
		return nil
	}
	ctx, cancel := ctxFromTimeoutMs(0)
	if args != nil {
		ctx, cancel = ctxFromTimeoutMs(args.TimeoutMs)
	}
	defer cancel()
	v, ok, err := s.Impl.BotPosition(ctx)
	if err != nil {
		return err
	}
	resp.Value, resp.OK = v, ok
	return nil
}

func (s *UQHolderModuleRPCServer) BotPositionOutOfSyncTick(args *UQHolderTimeoutArgs, resp *UQHolderInt64OKResp) error {
	if resp == nil {
		return nil
	}
	resp.Value, resp.OK = 0, false
	if s == nil || s.Impl == nil {
		return nil
	}
	ctx, cancel := ctxFromTimeoutMs(0)
	if args != nil {
		ctx, cancel = ctxFromTimeoutMs(args.TimeoutMs)
	}
	defer cancel()
	v, ok, err := s.Impl.BotPositionOutOfSyncTick(ctx)
	if err != nil {
		return err
	}
	resp.Value, resp.OK = v, ok
	return nil
}

func (s *UQHolderModuleRPCServer) ClientDimension(args *UQHolderTimeoutArgs, resp *UQHolderInt32OKResp) error {
	return s.callInt32OK(args, resp, func(ctx context.Context) (int32, bool, error) { return s.Impl.ClientDimension(ctx) })
}

func (s *UQHolderModuleRPCServer) ClientHotBarSlot(args *UQHolderTimeoutArgs, resp *UQHolderByteOKResp) error {
	if resp == nil {
		return nil
	}
	resp.Value, resp.OK = 0, false
	if s == nil || s.Impl == nil {
		return nil
	}
	ctx, cancel := ctxFromTimeoutMs(0)
	if args != nil {
		ctx, cancel = ctxFromTimeoutMs(args.TimeoutMs)
	}
	defer cancel()
	v, ok, err := s.Impl.ClientHotBarSlot(ctx)
	if err != nil {
		return err
	}
	resp.Value, resp.OK = uint8(v), ok
	return nil
}

func (s *UQHolderModuleRPCServer) ClientHoldingItem(args *UQHolderTimeoutArgs, resp *UQHolderMapOKResp) error {
	if resp == nil {
		return nil
	}
	resp.Value, resp.OK = nil, false
	if s == nil || s.Impl == nil {
		return nil
	}
	ctx, cancel := ctxFromTimeoutMs(0)
	if args != nil {
		ctx, cancel = ctxFromTimeoutMs(args.TimeoutMs)
	}
	defer cancel()
	v, ok, err := s.Impl.ClientHoldingItem(ctx)
	if err != nil {
		return err
	}
	resp.Value, resp.OK = v, ok
	return nil
}

func (s *UQHolderModuleRPCServer) GameRules(args *UQHolderTimeoutArgs, resp *UQHolderGameRulesResp) error {
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
	v, err := s.Impl.GameRules(ctx)
	if err != nil {
		return err
	}
	resp.Value = v
	return nil
}

func (s *UQHolderModuleRPCServer) ExtendRaw(args *UQHolderTimeoutArgs, resp *UQHolderMapResp) error {
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
	v, err := s.Impl.ExtendRaw(ctx)
	if err != nil {
		return err
	}
	resp.Value = v
	return nil
}

func (s *UQHolderModuleRPCServer) callString(args *UQHolderTimeoutArgs, resp *UQHolderStringResp, fn func(context.Context) (string, error)) error {
	if resp == nil {
		return nil
	}
	resp.Value = ""
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

func (s *UQHolderModuleRPCServer) callInt32OK(args *UQHolderTimeoutArgs, resp *UQHolderInt32OKResp, fn func(context.Context) (int32, bool, error)) error {
	if resp == nil {
		return nil
	}
	resp.Value, resp.OK = 0, false
	if s == nil || s.Impl == nil || fn == nil {
		return nil
	}
	ctx, cancel := ctxFromTimeoutMs(0)
	if args != nil {
		ctx, cancel = ctxFromTimeoutMs(args.TimeoutMs)
	}
	defer cancel()
	v, ok, err := fn(ctx)
	if err != nil {
		return err
	}
	resp.Value, resp.OK = v, ok
	return nil
}

func (s *UQHolderModuleRPCServer) callFloat32OK(args *UQHolderTimeoutArgs, resp *UQHolderFloat32OKResp, fn func(context.Context) (float32, bool, error)) error {
	if resp == nil {
		return nil
	}
	resp.Value, resp.OK = 0, false
	if s == nil || s.Impl == nil || fn == nil {
		return nil
	}
	ctx, cancel := ctxFromTimeoutMs(0)
	if args != nil {
		ctx, cancel = ctxFromTimeoutMs(args.TimeoutMs)
	}
	defer cancel()
	v, ok, err := fn(ctx)
	if err != nil {
		return err
	}
	resp.Value, resp.OK = v, ok
	return nil
}

type uqholderModuleRPCClient struct {
	c  *rpc.Client
	mu sync.Mutex
}

func newUQHolderModuleRPCClient(conn net.Conn) api.UQHolderModule {
	if conn == nil {
		return nil
	}
	return &uqholderModuleRPCClient{c: rpc.NewClient(conn)}
}

func (c *uqholderModuleRPCClient) Name() string { return api.NameUQHolderModule }

func (c *uqholderModuleRPCClient) BotName(ctx context.Context) (string, error) {
	return c.callString("Plugin.BotName", ctx)
}

func (c *uqholderModuleRPCClient) BotRuntimeID(ctx context.Context) (uint64, error) {
	var resp UQHolderUint64Resp
	err := c.call("Plugin.BotRuntimeID", ctx, &UQHolderTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}, &resp)
	return resp.Value, err
}

func (c *uqholderModuleRPCClient) BotUniqueID(ctx context.Context) (int64, error) {
	var resp UQHolderInt64Resp
	err := c.call("Plugin.BotUniqueID", ctx, &UQHolderTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}, &resp)
	return resp.Value, err
}

func (c *uqholderModuleRPCClient) BotIdentity(ctx context.Context) (string, error) {
	return c.callString("Plugin.BotIdentity", ctx)
}

func (c *uqholderModuleRPCClient) BotUUIDStr(ctx context.Context) (string, error) {
	return c.callString("Plugin.BotUUIDStr", ctx)
}

func (c *uqholderModuleRPCClient) BotXUID(ctx context.Context) (string, error) {
	return c.callString("Plugin.BotXUID", ctx)
}

func (c *uqholderModuleRPCClient) BasicRaw(ctx context.Context) (map[string]any, error) {
	var resp UQHolderMapResp
	err := c.call("Plugin.BasicRaw", ctx, &UQHolderTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}, &resp)
	return resp.Value, err
}

func (c *uqholderModuleRPCClient) CompressThreshold(ctx context.Context) (uint16, bool, error) {
	var resp UQHolderUint16OKResp
	err := c.call("Plugin.CompressThreshold", ctx, &UQHolderTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}, &resp)
	return resp.Value, resp.OK, err
}

func (c *uqholderModuleRPCClient) WorldGameMode(ctx context.Context) (int32, bool, error) {
	return c.callInt32OK("Plugin.WorldGameMode", ctx)
}

func (c *uqholderModuleRPCClient) GameMode(ctx context.Context) (int32, bool, error) {
	return c.callInt32OK("Plugin.GameMode", ctx)
}

func (c *uqholderModuleRPCClient) WorldDifficulty(ctx context.Context) (uint32, bool, error) {
	var resp UQHolderUint32OKResp
	err := c.call("Plugin.WorldDifficulty", ctx, &UQHolderTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}, &resp)
	return resp.Value, resp.OK, err
}

func (c *uqholderModuleRPCClient) Time(ctx context.Context) (int32, bool, error) {
	return c.callInt32OK("Plugin.Time", ctx)
}

func (c *uqholderModuleRPCClient) DayTime(ctx context.Context) (int32, bool, error) {
	return c.callInt32OK("Plugin.DayTime", ctx)
}

func (c *uqholderModuleRPCClient) DayTimePercent(ctx context.Context) (float32, bool, error) {
	var resp UQHolderFloat32OKResp
	err := c.call("Plugin.DayTimePercent", ctx, &UQHolderTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}, &resp)
	return resp.Value, resp.OK, err
}

func (c *uqholderModuleRPCClient) CurrentTick(ctx context.Context) (int64, bool, error) {
	var resp UQHolderInt64OKResp
	err := c.call("Plugin.CurrentTick", ctx, &UQHolderTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}, &resp)
	return resp.Value, resp.OK, err
}

func (c *uqholderModuleRPCClient) SyncRatio(ctx context.Context) (float32, bool, error) {
	var resp UQHolderFloat32OKResp
	err := c.call("Plugin.SyncRatio", ctx, &UQHolderTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}, &resp)
	return resp.Value, resp.OK, err
}

func (c *uqholderModuleRPCClient) BotDimension(ctx context.Context) (int32, bool, error) {
	return c.callInt32OK("Plugin.BotDimension", ctx)
}

func (c *uqholderModuleRPCClient) BotPosition(ctx context.Context) ([3]float32, bool, error) {
	var resp UQHolderPositionOKResp
	err := c.call("Plugin.BotPosition", ctx, &UQHolderTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}, &resp)
	return resp.Value, resp.OK, err
}

func (c *uqholderModuleRPCClient) BotPositionOutOfSyncTick(ctx context.Context) (int64, bool, error) {
	var resp UQHolderInt64OKResp
	err := c.call("Plugin.BotPositionOutOfSyncTick", ctx, &UQHolderTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}, &resp)
	return resp.Value, resp.OK, err
}

func (c *uqholderModuleRPCClient) ClientDimension(ctx context.Context) (int32, bool, error) {
	return c.callInt32OK("Plugin.ClientDimension", ctx)
}

func (c *uqholderModuleRPCClient) ClientHotBarSlot(ctx context.Context) (byte, bool, error) {
	var resp UQHolderByteOKResp
	err := c.call("Plugin.ClientHotBarSlot", ctx, &UQHolderTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}, &resp)
	return byte(resp.Value), resp.OK, err
}

func (c *uqholderModuleRPCClient) ClientHoldingItem(ctx context.Context) (map[string]any, bool, error) {
	var resp UQHolderMapOKResp
	err := c.call("Plugin.ClientHoldingItem", ctx, &UQHolderTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}, &resp)
	return resp.Value, resp.OK, err
}

func (c *uqholderModuleRPCClient) GameRules(ctx context.Context) (map[string]api.GameRule, error) {
	var resp UQHolderGameRulesResp
	err := c.call("Plugin.GameRules", ctx, &UQHolderTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}, &resp)
	return resp.Value, err
}

func (c *uqholderModuleRPCClient) ExtendRaw(ctx context.Context) (map[string]any, error) {
	var resp UQHolderMapResp
	err := c.call("Plugin.ExtendRaw", ctx, &UQHolderTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}, &resp)
	return resp.Value, err
}

func (c *uqholderModuleRPCClient) callString(method string, ctx context.Context) (string, error) {
	var resp UQHolderStringResp
	err := c.call(method, ctx, &UQHolderTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}, &resp)
	return resp.Value, err
}

func (c *uqholderModuleRPCClient) callInt32OK(method string, ctx context.Context) (int32, bool, error) {
	var resp UQHolderInt32OKResp
	err := c.call(method, ctx, &UQHolderTimeoutArgs{TimeoutMs: timeoutMsFromContext(ctx)}, &resp)
	return resp.Value, resp.OK, err
}

func (c *uqholderModuleRPCClient) call(method string, _ context.Context, args any, resp any) error {
	if c == nil || c.c == nil {
		return errors.New("uqholderModuleRPCClient: client is not initialised")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call(method, args, resp)
}

var _ api.UQHolderModule = (*uqholderModuleRPCClient)(nil)
