package protocol

import (
	"errors"
	"net"
	"net/rpc"
	"strings"
	"sync"

	"github.com/hashicorp/go-plugin"

	"github.com/Yeah114/tempest-plugin-sdk/api"
)

type DatabaseModuleNameResp struct {
	Name string
}

type DatabaseModuleKeyValueDBArgs struct {
	Name   string
	DBType string
}

type DatabaseModuleKeyValueDBResp struct {
	Exists      bool
	DBBrokerID  uint32
	ModuleError string
}

type DatabaseModuleRPCServer struct {
	Impl   api.DatabaseModule
	broker *plugin.MuxBroker
}

func (s *DatabaseModuleRPCServer) Name(_ *Empty, resp *DatabaseModuleNameResp) error {
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

func (s *DatabaseModuleRPCServer) KeyValueDB(args *DatabaseModuleKeyValueDBArgs, resp *DatabaseModuleKeyValueDBResp) error {
	if resp == nil {
		return nil
	}
	resp.Exists = false
	resp.DBBrokerID = 0
	resp.ModuleError = ""
	if s == nil || s.Impl == nil || args == nil {
		return nil
	}
	if s.broker == nil {
		return errors.New("DatabaseModuleRPCServer.KeyValueDB: broker unavailable")
	}

	name := strings.TrimSpace(args.Name)
	dbType := strings.TrimSpace(args.DBType)
	db, err := s.Impl.KeyValueDB(name, dbType)
	if err != nil {
		return err
	}
	if db == nil {
		return nil
	}

	id := s.broker.NextId()
	go acceptAndServeMuxBroker(s.broker, id, &KeyValueDBRPCServer{Impl: db, broker: s.broker})
	resp.Exists = true
	resp.DBBrokerID = id
	return nil
}

type databaseModuleRPCClient struct {
	c      *rpc.Client
	broker *plugin.MuxBroker
	mu     sync.Mutex
}

func newDatabaseModuleRPCClient(conn net.Conn, broker *plugin.MuxBroker) api.DatabaseModule {
	if conn == nil {
		return nil
	}
	return &databaseModuleRPCClient{c: rpc.NewClient(conn), broker: broker}
}

func (c *databaseModuleRPCClient) Name() string { return api.NameDatabaseModule }

func (c *databaseModuleRPCClient) KeyValueDB(name string, dbType string) (api.KeyValueDB, error) {
	if c == nil || c.c == nil {
		return nil, errors.New("databaseModuleRPCClient: client is not initialised")
	}
	if c.broker == nil {
		return nil, errors.New("databaseModuleRPCClient: broker unavailable")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	var resp DatabaseModuleKeyValueDBResp
	if err := c.c.Call("Plugin.KeyValueDB", &DatabaseModuleKeyValueDBArgs{Name: name, DBType: dbType}, &resp); err != nil {
		return nil, err
	}
	if !resp.Exists || resp.DBBrokerID == 0 {
		return nil, nil
	}

	conn, err := c.broker.Dial(resp.DBBrokerID)
	if err != nil || conn == nil {
		return nil, err
	}
	return newKeyValueDBRPCClient(conn, c.broker), nil
}
