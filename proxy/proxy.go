package proxy

import (
	codeharvest "github.com/creativecreature/code-harvest"
)

// Server is the interface that the RPC server must satisfy.
type Server interface {
	FocusGained(event codeharvest.Event, reply *string)
	OpenFile(event codeharvest.Event, reply *string)
	SendHeartbeat(event codeharvest.Event, reply *string)
	EndSession(event codeharvest.Event, reply *string)
}

// Proxy serves as the intermediary between our client and server. It directs
// remote procedure calls to the server, mitigating the risk of unintentionally
// revealing server methods, just because they happen to conform to the RPC interface.
type Proxy struct {
	server Server
}

// New returns a new proxy.
func New(server Server) *Proxy {
	return &Proxy{server: server}
}

// FocusGained should be called when a buffer gains focus.
func (p *Proxy) FocusGained(event codeharvest.Event, reply *string) error {
	p.server.FocusGained(event, reply)
	return nil
}

// OpenFile should be called when a buffer is opened.
// The server will check if the path is a valid file.
func (p *Proxy) OpenFile(event codeharvest.Event, reply *string) error {
	p.server.OpenFile(event, reply)
	return nil
}

// SendHeartbeat can be called for events such as buffer writes
// and cursor moves. Its purpose is to notify the server that
// the current session remains active. If we don't perform any
// actions for 10 minutes the server is going to end the session.
func (p *Proxy) SendHeartbeat(event codeharvest.Event, reply *string) error {
	p.server.SendHeartbeat(event, reply)
	return nil
}

// EndSession should be called when the neovim process ends.
func (p *Proxy) EndSession(event codeharvest.Event, reply *string) error {
	p.server.EndSession(event, reply)
	return nil
}
