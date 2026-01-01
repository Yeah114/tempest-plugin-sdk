package define

type Frame interface {
	PluginFrame

	ListModules() map[string]Module
	GetModule(name string) (Module, bool)
}
