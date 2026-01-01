package protocol

import (
	"github.com/hashicorp/go-plugin"

	"github.com/Yeah114/tempest/tempest-plugin-sdk/api"
)

// Serve starts a go-plugin server for the provided plugin implementation.
// This is intended to be called from a standalone plugin binary.
func Serve(p api.Plugin) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins: map[string]plugin.Plugin{
			PluginKey: &DynamicRPCPlugin{Impl: p},
		},
	})
}
