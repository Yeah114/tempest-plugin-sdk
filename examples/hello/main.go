package main

import (
	"context"
	"fmt"

	"github.com/Yeah114/tempest/tempest-plugin-sdk/api"
	"github.com/Yeah114/tempest/tempest-plugin-sdk/protocol"
)

type HelloPlugin struct {
	api.BasicPlugin
}

func (p *HelloPlugin) Load(_ context.Context) error {
	fmt.Println("hello")
	return nil
}

func main() {
	protocol.Serve(&HelloPlugin{})
}
