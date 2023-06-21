package main

import (
	"code-harvest.conner.dev/client"
	"github.com/neovim/go-client/nvim/plugin"
)

// ldflags
var (
	serverName string
	port       string
	hostname   string
)

func main() {
	client, err := client.New(serverName, port, hostname)
	if err != nil {
		panic(err)
	}

	plugin.Main(func(p *plugin.Plugin) error {
		p.HandleFunction(&plugin.FunctionOptions{Name: "OnFocusGained"}, client.FocusGained)
		p.HandleFunction(&plugin.FunctionOptions{Name: "OpenFile"}, client.OpenFile)
		p.HandleFunction(&plugin.FunctionOptions{Name: "SendHeartbeat"}, client.SendHeartbeat)
		p.HandleFunction(&plugin.FunctionOptions{Name: "EndSession"}, client.EndSession)
		return nil
	})
}
