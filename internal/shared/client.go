package shared

import (
	"fmt"
	"net/rpc"
	"runtime"
)

// Client for making remote procedure calls to the server
type Client struct {
	rpcClient *rpc.Client
}

func NewClient(port, hostname string) (*Client, error) {
	rpcClient, err := rpc.DialHTTP("tcp", fmt.Sprintf("%s:%s", hostname, port))
	if err != nil {
		return nil, err
	}

	return &Client{rpcClient: rpcClient}, nil
}

// Should be invoked when a buffer gains focus
func (c *Client) FocusGained(args []string) {
	event := Event{Id: args[0], Path: args[1], Editor: "nvim", OS: runtime.GOOS}
	reply := ""
	serviceMethod := ServerName + ".FocusGained"
	c.rpcClient.Call(serviceMethod, event, &reply)
}

// Should be invoked when a file is opened
func (c *Client) OpenFile(args []string) {
	event := Event{Id: args[0], Path: args[1], Editor: "nvim", OS: runtime.GOOS}
	reply := ""
	serviceMethod := ServerName + ".OpenFile"
	c.rpcClient.Call(serviceMethod, event, &reply)
}

// Can be sent on writes, cursor moves, searches, etc. It lets the server know
// that the session is still active. If we don't perform any actions for 10
// minutes the server is going to end the session.
func (c *Client) SendHeartbeat(args []string) {
	event := Event{Id: args[0], Path: args[1], Editor: "nvim", OS: runtime.GOOS}
	reply := ""
	serviceMethod := ServerName + ".SendHeartbeat"
	c.rpcClient.Call(serviceMethod, event, &reply)
}

// Should be invoked when the editor closes.
func (c *Client) EndSession(args []string) {
	event := Event{Id: args[0], Path: args[1], Editor: "nvim", OS: runtime.GOOS}
	reply := ""
	serviceMethod := ServerName + ".EndSession"
	c.rpcClient.Call(serviceMethod, event, &reply)
}
