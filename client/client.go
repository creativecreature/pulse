// Package client is used to send remote procedure calls to the server.
package client

import (
	"fmt"
	"net/rpc"
	"runtime"

	"github.com/viccon/pulse"
)

// Client for making remote procedure calls to the server.
type Client struct {
	serverName string
	rpcClient  *rpc.Client
}

// createEvents creates a new event from the slice of arguments
// that we receive from the neovim client.
func createEvent(args []string) pulse.Event {
	filetype := args[2]
	if filetype == "typescript.tsx" {
		filetype = "typescript"
	}

	return pulse.Event{
		EditorID: args[0],
		Path:     args[1],
		Filetype: filetype,
		Editor:   "nvim",
		OS:       runtime.GOOS,
	}
}

// New is used to create a new client.
func New(serverName, port, hostname string) (*Client, error) {
	rpcClient, err := rpc.DialHTTP("tcp", fmt.Sprintf("%s:%s", hostname, port))
	if err != nil {
		return nil, err
	}

	return &Client{serverName: serverName, rpcClient: rpcClient}, nil
}

// FocusGained should be called when a buffer gains focus.
func (c *Client) FocusGained(args []string) {
	event, reply := createEvent(args), ""
	serviceMethod := c.serverName + ".FocusGained"
	//nolint: errcheck // I don't want to print eventual errors in the editor.
	c.rpcClient.Call(serviceMethod, event, &reply)
}

// OpenFile should be called when a buffer is opened. The server
// will check if the path is a valid file.
func (c *Client) OpenFile(args []string) {
	event, reply := createEvent(args), ""
	serviceMethod := c.serverName + ".OpenFile"
	//nolint: errcheck // I don't want to print eventual errors in the editor.
	c.rpcClient.Call(serviceMethod, event, &reply)
}

// SendHeartbeat can be called for events such as buffer writes and cursor moves.
// Its purpose is to notify the server that the current session remains active.
// The server ends the session if we don't perform any actions for 10 minutes.
func (c *Client) SendHeartbeat(args []string) {
	event, reply := createEvent(args), ""
	serviceMethod := c.serverName + ".SendHeartbeat"
	//nolint: errcheck // I don't want to print eventual errors in the editor.
	c.rpcClient.Call(serviceMethod, event, &reply)
}

// EndSession should be called when the neovim process ends.
func (c *Client) EndSession(args []string) {
	event, reply := createEvent(args), ""
	serviceMethod := c.serverName + ".EndSession"
	//nolint: errcheck // I don't want to print eventual errors in the editor.
	c.rpcClient.Call(serviceMethod, event, &reply)
}
