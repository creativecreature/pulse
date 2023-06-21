package client

import (
	"fmt"
	"net/rpc"
	"runtime"

	"code-harvest.conner.dev/domain"
)

// Client for making remote procedure calls to the server
type Client struct {
	serverName string
	rpcClient  *rpc.Client
}

// createEvents creates a new event from the arguments we receive
func createEvent(args []string) domain.Event {
	return domain.Event{
		Id:     args[0],
		Path:   args[1],
		Editor: "nvim",
		OS:     runtime.GOOS,
	}
}

// New is used to create a new client
func New(serverName, port, hostname string) (*Client, error) {
	rpcClient, err := rpc.DialHTTP("tcp", fmt.Sprintf("%s:%s", hostname, port))
	if err != nil {
		return nil, err
	}

	return &Client{serverName: serverName, rpcClient: rpcClient}, nil
}

// FocusGained should be called when a buffer gains focus
func (c *Client) FocusGained(args []string) {
	event, reply := createEvent(args), ""
	serviceMethod := c.serverName + ".FocusGained"
	c.rpcClient.Call(serviceMethod, event, &reply)
}

// OpenFile should be called when a file is opened
func (c *Client) OpenFile(args []string) {
	event, reply := createEvent(args), ""
	serviceMethod := c.serverName + ".OpenFile"
	c.rpcClient.Call(serviceMethod, event, &reply)
}

// SendHeartbeat can be called on writes, cursor moves, searches, etc. It lets
// the server know that the session is still active. If we don't perform any
// actions for 10 minutes the server is going to end the session
func (c *Client) SendHeartbeat(args []string) {
	event, reply := createEvent(args), ""
	serviceMethod := c.serverName + ".SendHeartbeat"
	c.rpcClient.Call(serviceMethod, event, &reply)
}

// EndSession should be called when the editor closes
func (c *Client) EndSession(args []string) {
	event, reply := createEvent(args), ""
	serviceMethod := c.serverName + ".EndSession"
	c.rpcClient.Call(serviceMethod, event, &reply)
}
