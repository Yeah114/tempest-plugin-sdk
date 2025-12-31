package protocol

import (
	"github.com/hashicorp/go-plugin"
)

const PluginKey = "tempest_dynamic_v1"

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "TEMPEST_DYNAMIC_PLUGIN",
	MagicCookieValue: "1",
}
