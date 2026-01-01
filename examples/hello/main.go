package main

import (
	"context"

	"github.com/Yeah114/tempest-plugin-sdk/api"
	"github.com/Yeah114/tempest-plugin-sdk/protocol"
)

type HelloPlugin struct {
	api.BasicPlugin
}

func (p *HelloPlugin) Load(_ context.Context) error {
	terminal, _ := api.GetModule[api.TerminalModule](p.Frame(), api.NameTerminalModule)
	terminal.Success("hello", "hello")
	terminalMenu, _ := api.GetModule[api.TerminalMenuModule](p.Frame(), api.NameTerminalMenuModule)
	_ = terminalMenu.RegisterMenuEntry(&api.TerminalMenuEntry{
		Triggers: []string{"hello"},
		ArgumentHint: "hello",
		Usage: "hello",
		OnTrigger: func(_ []string) {
			terminal.Info("hello", "hello")
		},
	})
	return nil
}

func main() {
	protocol.Serve(&HelloPlugin{})
}
