package main

import (
	"code-harvest.conner.dev/internal/shared"
	"github.com/neovim/go-client/nvim/plugin"
)

// These are set by linker flags.
var (
	port     string
	hostname string
)

func main() {
	client, err := shared.NewClient(port, hostname)
	if err != nil {
		panic(err)
	}
	// Add these functions to NVIM so that I can map them to autocommands.
	plugin.Main(func(p *plugin.Plugin) error {
		p.HandleFunction(&plugin.FunctionOptions{Name: "OnFocusGained"}, client.FocusGained)
		p.HandleFunction(&plugin.FunctionOptions{Name: "OpenFile"}, client.OpenFile)
		p.HandleFunction(&plugin.FunctionOptions{Name: "SendHeartbeat"}, client.SendHeartbeat)
		p.HandleFunction(&plugin.FunctionOptions{Name: "EndSession"}, client.EndSession)
		return nil
	})
}
