package define

// PluginFrame exposes host plugin management capabilities to remote plugins.
// A remote plugin can type-assert Frame to this interface.
type PluginFrame interface {
	GetPluginConfig(id string) (PluginConfig, bool)

	// Activate event is triggered after the framework finishes loading all plugins.
	RegisterWhenActivate(handler func()) (string, error)
	UnregisterWhenActivate(listenerID string) bool
}
