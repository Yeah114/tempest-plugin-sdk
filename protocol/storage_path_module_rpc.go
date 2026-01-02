package protocol

import (
	"errors"
	"net"
	"net/rpc"
	"sync"

	"github.com/Yeah114/tempest-plugin-sdk/api"
)

type StoragePathModuleNameResp struct {
	Name string
}

type StoragePathArgs struct {
	Parts []string
}

type StoragePathResp struct {
	Path string
}

type StoragePathModuleRPCServer struct {
	Impl api.StoragePathModule
}

func (s *StoragePathModuleRPCServer) Name(_ *Empty, resp *StoragePathModuleNameResp) error {
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

func (s *StoragePathModuleRPCServer) ConfigPath(args *StoragePathArgs, resp *StoragePathResp) error {
	if resp == nil {
		return nil
	}
	resp.Path = ""
	if s == nil || s.Impl == nil {
		return nil
	}
	var parts []string
	if args != nil && args.Parts != nil {
		parts = args.Parts
	}
	resp.Path = s.Impl.ConfigPath(parts...)
	return nil
}

func (s *StoragePathModuleRPCServer) CodePath(args *StoragePathArgs, resp *StoragePathResp) error {
	if resp == nil {
		return nil
	}
	resp.Path = ""
	if s == nil || s.Impl == nil {
		return nil
	}
	var parts []string
	if args != nil && args.Parts != nil {
		parts = args.Parts
	}
	resp.Path = s.Impl.CodePath(parts...)
	return nil
}

func (s *StoragePathModuleRPCServer) DataFilePath(args *StoragePathArgs, resp *StoragePathResp) error {
	if resp == nil {
		return nil
	}
	resp.Path = ""
	if s == nil || s.Impl == nil {
		return nil
	}
	var parts []string
	if args != nil && args.Parts != nil {
		parts = args.Parts
	}
	resp.Path = s.Impl.DataFilePath(parts...)
	return nil
}

func (s *StoragePathModuleRPCServer) CachePath(args *StoragePathArgs, resp *StoragePathResp) error {
	if resp == nil {
		return nil
	}
	resp.Path = ""
	if s == nil || s.Impl == nil {
		return nil
	}
	var parts []string
	if args != nil && args.Parts != nil {
		parts = args.Parts
	}
	resp.Path = s.Impl.CachePath(parts...)
	return nil
}

type storagePathModuleRPCClient struct {
	c  *rpc.Client
	mu sync.Mutex
}

func newStoragePathModuleRPCClient(conn net.Conn) api.StoragePathModule {
	if conn == nil {
		return nil
	}
	return &storagePathModuleRPCClient{c: rpc.NewClient(conn)}
}

func (c *storagePathModuleRPCClient) Name() string { return api.NameStoragePathModule }

func (c *storagePathModuleRPCClient) call(method string, parts []string) (string, error) {
	if c == nil || c.c == nil {
		return "", errors.New("storagePathModuleRPCClient: client is not initialised")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var resp StoragePathResp
	if err := c.c.Call(method, &StoragePathArgs{Parts: parts}, &resp); err != nil {
		return "", err
	}
	return resp.Path, nil
}

func (c *storagePathModuleRPCClient) ConfigPath(parts ...string) string {
	out, _ := c.call("Plugin.ConfigPath", parts)
	return out
}
func (c *storagePathModuleRPCClient) CodePath(parts ...string) string {
	out, _ := c.call("Plugin.CodePath", parts)
	return out
}
func (c *storagePathModuleRPCClient) DataFilePath(parts ...string) string {
	out, _ := c.call("Plugin.DataFilePath", parts)
	return out
}
func (c *storagePathModuleRPCClient) CachePath(parts ...string) string {
	out, _ := c.call("Plugin.CachePath", parts)
	return out
}
