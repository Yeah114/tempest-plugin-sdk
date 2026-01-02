package protocol

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

// acceptAndServeMuxBroker is a quiet version of go-plugin's (*MuxBroker).AcceptAndServe.
// The upstream helper logs on timeout; we intentionally swallow the timeout because it can happen
// during frame/plugin restarts and shouldn't be treated as a hard error by plugin authors.
func acceptAndServeMuxBroker(broker *plugin.MuxBroker, id uint32, v interface{}) {
	if broker == nil || id == 0 || v == nil {
		return
	}
	conn, err := broker.Accept(id)
	if err != nil || conn == nil {
		return
	}
	defer func() { _ = conn.Close() }()

	srv := rpc.NewServer()
	if err := srv.RegisterName("Plugin", v); err != nil {
		return
	}
	srv.ServeConn(conn)
}
