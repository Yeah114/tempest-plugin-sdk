package define

type Daemon interface {
	Name() (name string)
	ReConfig(config map[string]interface{}) (err error)
	Config() (config map[string]interface{})
}
