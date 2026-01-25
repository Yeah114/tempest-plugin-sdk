package api

import "github.com/Yeah114/EmptyDea-plugin-sdk/define"

func GetModule[T any](frame define.Frame, moduleName string) (module T, ok bool) {
	mod, ok := frame.GetModule(moduleName)
	if !ok {
		return
	}
	module, ok = mod.(T)
	return
}
