package main

import (
	"github.com/viccon/pulse"
	"github.com/viccon/pulse/client"
	"github.com/neovim/go-client/nvim/plugin"
)

func main() {
	cfg, err := pulse.ParseConfig()
	if err != nil {
		panic("failed to parse config")
	}

	client, err := client.New(cfg.Server.Name, cfg.Server.Port, cfg.Server.Hostname)
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
