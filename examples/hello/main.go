package main

import (
	"context"
	"fmt"

	"github.com/Yeah114/tempest-plugin-sdk/api"
	"github.com/Yeah114/tempest-plugin-sdk/protocol"
	"github.com/mitchellh/mapstructure"
)

type HelloPluginConfig struct {
	UserName string `mapstructure:"名字"`
}

type HelloPlugin struct {
	HelloPluginConfig
	api.BasicPlugin
	api.PluginTool
	api.TerminalMenuModule
}

func (p *HelloPlugin) Load(_ context.Context) (err error) {
	err = mapstructure.Decode(p.Config(), &p.HelloPluginConfig)
	if err != nil {
		return fmt.Errorf("HelloPlugin.Load: 解析插件配置时发生错误: %v", err)
	}

	p.PluginTool = api.NewPluginTool(p)
	p.TerminalMenuModule, _ = api.GetModule[api.TerminalMenuModule](p.Frame(), api.NameTerminalMenuModule)

	_ = p.RegisterMenuEntry(&api.TerminalMenuEntry{
		Triggers: []string{"hello"},
		Usage: "打个招呼",
		OnTrigger: func(_ []string) {
			p.Info(fmt.Sprintf("hello, %s!", p.UserName))
		},
	})

	return nil
}

func main() {
	protocol.Serve(&HelloPlugin{})
}
