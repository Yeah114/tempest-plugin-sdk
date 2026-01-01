package protocol

import (
	"sync"

	"github.com/hashicorp/go-plugin"

	"github.com/Yeah114/tempest-plugin-sdk/api"
)

var (
	serveMu          sync.RWMutex
	serveTestConfig  *plugin.ServeTestConfig
)

// SetServeTestConfig enables go-plugin "test mode" for subsequent Serve calls.
// It is used by the host when running plugins in-process (e.g. via yaegi).
// restore resets the previous config.
func SetServeTestConfig(cfg *plugin.ServeTestConfig) (restore func()) {
	serveMu.Lock()
	prev := serveTestConfig
	serveTestConfig = cfg
	serveMu.Unlock()
	return func() {
		serveMu.Lock()
		serveTestConfig = prev
		serveMu.Unlock()
	}
}

func currentServeTestConfig() *plugin.ServeTestConfig {
	serveMu.RLock()
	cfg := serveTestConfig
	serveMu.RUnlock()
	return cfg
}

// Serve starts a go-plugin server for the provided plugin implementation.
// This is intended to be called from a standalone plugin binary.
func Serve(p api.Plugin) {
	cfg := &plugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins: map[string]plugin.Plugin{
			PluginKey: &DynamicRPCPlugin{Impl: p},
		},
	}
	if tc := currentServeTestConfig(); tc != nil {
		cfg.Test = tc
	}
	plugin.Serve(cfg)
}
