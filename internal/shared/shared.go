package shared

import (
	"fmt"
	"net/rpc"
	"runtime"
)

var ServerName = "CodeHarvestApp"

type Event struct {
	Id     string
	Path   string
	Editor string
	OS     string
}

type Server interface {
	FocusGained(event Event, reply *string) error
	OpenFile(event Event, reply *string) error
	SendHeartbeat(event Event, reply *string) error
	EndSession(event Event, reply *string) error
}

type Proxy struct {
	server Server
}

func NewProxy(server Server) *Proxy {
	return &Proxy{server: server}
}

func (p *Proxy) FocusGained(event Event, reply *string) error {
	return p.server.FocusGained(event, reply)
}

func (p *Proxy) OpenFile(event Event, reply *string) error {
	return p.server.OpenFile(event, reply)
}

func (p *Proxy) SendHeartbeat(event Event, reply *string) error {
	return p.server.SendHeartbeat(event, reply)
}

func (p *Proxy) EndSession(event Event, reply *string) error {
	return p.server.EndSession(event, reply)
}

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

func (c *Client) FocusGained(args []string) {
	event := Event{Id: args[0], Path: args[1], Editor: "nvim", OS: runtime.GOOS}
	reply := ""
	serviceMethod := ServerName + ".FocusGained"
	c.rpcClient.Call(serviceMethod, event, &reply)
}

func (c *Client) OpenFile(args []string) {
	event := Event{Id: args[0], Path: args[1], Editor: "nvim", OS: runtime.GOOS}
	reply := ""
	serviceMethod := ServerName + ".OpenFile"
	c.rpcClient.Call(serviceMethod, event, &reply)
}

func (c *Client) SendHeartbeat(args []string) {
	event := Event{Id: args[0], Path: args[1], Editor: "nvim", OS: runtime.GOOS}
	reply := ""
	serviceMethod := ServerName + ".SendHeartbeat"
	c.rpcClient.Call(serviceMethod, event, &reply)
}

func (c *Client) EndSession(args []string) {
	event := Event{Id: args[0], Path: args[1], Editor: "nvim", OS: runtime.GOOS}
	reply := ""
	serviceMethod := ServerName + ".EndSession"
	c.rpcClient.Call(serviceMethod, event, &reply)
}
