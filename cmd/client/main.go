package main

import (
	"fmt"
	"log"
	"net/rpc"
	"runtime"

	"code-harvest.conner.dev/internal/shared"
	"github.com/neovim/go-client/nvim/plugin"
)

// These are set by linker flags.
var port string
var hostname string

// Get a rpc client that is connected to the server.
func getClient() *rpc.Client {
	client, err := rpc.DialHTTP("tcp", fmt.Sprintf("%s:%s", hostname, port))
	if err != nil {
		log.Printf("An error occured when we tried to connect :%s\n", err)
	}

	return client
}

type CodeHarvestClient struct {
	rpcClient *rpc.Client
}

func (c *CodeHarvestClient) focusGained(args []string) {
	focusGainedArgs := struct {
		Id     string
		Path   string
		Editor string
		OS     string
	}{Id: args[0], Path: args[1], Editor: "nvim", OS: runtime.GOOS}

	reply := ""
	serviceMethod := shared.ServerName + ".FocusGained"
	err := c.rpcClient.Call(serviceMethod, focusGainedArgs, &reply)
	if err != nil {
		log.Printf("An error occured when calling FocusGained :%s\n", err)
	}
}

func (c *CodeHarvestClient) openFile(args []string) {
	openFileArgs := struct {
		Id     string
		Path   string
		Editor string
		OS     string
	}{Id: args[0], Path: args[1], Editor: "nvim", OS: runtime.GOOS}

	reply := ""
	serviceMethod := shared.ServerName + ".OpenFile"
	err := c.rpcClient.Call(serviceMethod, openFileArgs, &reply)
	if err != nil {
		log.Printf("An error occured when calling OpenFile :%s\n", err)
	}
}

func (c *CodeHarvestClient) sendHeartbeat(args []string) {
	heartbeatArgs := struct {
		Id     string
		Path   string
		Editor string
		OS     string
	}{Id: args[0], Path: args[1], Editor: "nvim", OS: runtime.GOOS}

	reply := ""
	serviceMethod := shared.ServerName + ".SendHeartbeat"
	err := c.rpcClient.Call(serviceMethod, heartbeatArgs, &reply)
	if err != nil {
		log.Printf("An error occured when calling SendHeartbeat :%s\n", err)
	}
}

func (c *CodeHarvestClient) endSession(args []string) {
	endSessionArgs := struct {
		Id string
	}{Id: args[0]}

	reply := ""
	serviceMethod := shared.ServerName + ".EndSession"
	err := c.rpcClient.Call(serviceMethod, endSessionArgs, &reply)

	if err != nil {
		log.Fatal("An error occurred when calling EndSession", err)
	}
}

func main() {
	client := &CodeHarvestClient{
		rpcClient: getClient(),
	}

	// Add these functions to NVIM so that I can map them to autocommands.
	plugin.Main(func(p *plugin.Plugin) error {
		p.HandleFunction(&plugin.FunctionOptions{Name: "OnFocusGained"}, client.focusGained)
		p.HandleFunction(&plugin.FunctionOptions{Name: "OpenFile"}, client.openFile)
		p.HandleFunction(&plugin.FunctionOptions{Name: "SendHeartbeat"}, client.sendHeartbeat)
		p.HandleFunction(&plugin.FunctionOptions{Name: "EndSession"}, client.endSession)
		return nil
	})
}
