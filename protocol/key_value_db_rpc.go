package protocol

import (
	"errors"
	"net"
	"net/rpc"
	"sync"

	"github.com/hashicorp/go-plugin"

	"github.com/Yeah114/EmptyDea-plugin-sdk/api"
)

type KeyValueDBGetArgs struct {
	Key string
}

type KeyValueDBGetResp struct {
	Value string
	OK    bool
}

type KeyValueDBSetArgs struct {
	Key   string
	Value string
}

type KeyValueDBDeleteArgs struct {
	Key string
}

type KeyValueDBIterateArgs struct {
	CallbackBrokerID uint32
}

type KeyValueDBIterateEntryArgs struct {
	Key   string
	Value string
}

type KeyValueDBRPCServer struct {
	Impl   api.KeyValueDB
	broker *plugin.MuxBroker
}

func (s *KeyValueDBRPCServer) Get(args *KeyValueDBGetArgs, resp *KeyValueDBGetResp) error {
	if resp == nil {
		return nil
	}
	resp.Value = ""
	resp.OK = false
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	v, ok, err := s.Impl.Get(args.Key)
	if err != nil {
		return err
	}
	resp.Value = v
	resp.OK = ok
	return nil
}

func (s *KeyValueDBRPCServer) Set(args *KeyValueDBSetArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	return s.Impl.Set(args.Key, args.Value)
}

func (s *KeyValueDBRPCServer) Delete(args *KeyValueDBDeleteArgs, _ *Empty) error {
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	return s.Impl.Delete(args.Key)
}

func (s *KeyValueDBRPCServer) Iterate(args *KeyValueDBIterateArgs, resp *BoolResp) error {
	if resp == nil {
		return nil
	}
	resp.OK = false
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	if s.broker == nil || args.CallbackBrokerID == 0 {
		return errors.New("KeyValueDBRPCServer.Iterate: callback broker id is 0")
	}

	conn, err := s.broker.Dial(args.CallbackBrokerID)
	if err != nil || conn == nil {
		return err
	}
	client := rpc.NewClient(conn)
	defer client.Close()

	err = s.Impl.Iterate(func(key, value string) bool {
		var out BoolResp
		callErr := client.Call("Plugin.OnEntry", &KeyValueDBIterateEntryArgs{Key: key, Value: value}, &out)
		if callErr != nil {
			return false
		}
		return out.OK
	})
	if err != nil {
		return err
	}
	resp.OK = true
	return nil
}

type keyValueDBIterateCallbackRPCServer struct {
	Handler func(key, value string) bool
}

func (s *keyValueDBIterateCallbackRPCServer) OnEntry(args *KeyValueDBIterateEntryArgs, resp *BoolResp) error {
	if resp == nil {
		return nil
	}
	resp.OK = false
	if s == nil || s.Handler == nil || args == nil {
		return nil
	}
	resp.OK = s.Handler(args.Key, args.Value)
	return nil
}

type keyValueDBRPCClient struct {
	c      *rpc.Client
	broker *plugin.MuxBroker
	mu     sync.Mutex
}

func newKeyValueDBRPCClient(conn net.Conn, broker *plugin.MuxBroker) api.KeyValueDB {
	if conn == nil {
		return nil
	}
	return &keyValueDBRPCClient{c: rpc.NewClient(conn), broker: broker}
}

func (c *keyValueDBRPCClient) Get(key string) (string, bool, error) {
	if c == nil || c.c == nil {
		return "", false, errors.New("keyValueDBRPCClient: client is not initialised")
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	var resp KeyValueDBGetResp
	if err := c.c.Call("Plugin.Get", &KeyValueDBGetArgs{Key: key}, &resp); err != nil {
		return "", false, err
	}
	return resp.Value, resp.OK, nil
}

func (c *keyValueDBRPCClient) Set(key, value string) error {
	if c == nil || c.c == nil {
		return errors.New("keyValueDBRPCClient: client is not initialised")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.Set", &KeyValueDBSetArgs{Key: key, Value: value}, &Empty{})
}

func (c *keyValueDBRPCClient) Delete(key string) error {
	if c == nil || c.c == nil {
		return errors.New("keyValueDBRPCClient: client is not initialised")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Call("Plugin.Delete", &KeyValueDBDeleteArgs{Key: key}, &Empty{})
}

func (c *keyValueDBRPCClient) Iterate(fn func(key, value string) bool) error {
	if c == nil || c.c == nil {
		return errors.New("keyValueDBRPCClient: client is not initialised")
	}
	if fn == nil {
		return nil
	}
	if c.broker == nil {
		return errors.New("keyValueDBRPCClient.Iterate: broker unavailable")
	}

	callbackID := c.broker.NextId()
	go acceptAndServeMuxBroker(c.broker, callbackID, &keyValueDBIterateCallbackRPCServer{Handler: fn})

	c.mu.Lock()
	defer c.mu.Unlock()
	var resp BoolResp
	return c.c.Call("Plugin.Iterate", &KeyValueDBIterateArgs{CallbackBrokerID: callbackID}, &resp)
}

func (c *keyValueDBRPCClient) MigrateTo(dst api.KeyValueDB) error {
	if dst == nil {
		return errors.New("keyValueDBRPCClient.MigrateTo: dst is nil")
	}
	var migrateErr error
	_ = c.Iterate(func(key, value string) bool {
		if err := dst.Set(key, value); err != nil {
			migrateErr = err
			return false
		}
		return true
	})
	return migrateErr
}

func (c *keyValueDBRPCClient) Close() error {
	if c == nil || c.c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.c.Close()
}
