package define

type Frame interface {
	ListModules() map[string]Module
	GetModule(name string) (Module, bool)
}
