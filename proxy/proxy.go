package proxy

import "code-harvest.conner.dev/domain"

// Server is the interface that the RPC server must satisfy
type Server interface {
	FocusGained(event domain.Event, reply *string) error
	OpenFile(event domain.Event, reply *string) error
	SendHeartbeat(event domain.Event, reply *string) error
	EndSession(event domain.Event, reply *string) error
}

// Proxy is the layer between our client and server. It forwards remote
// procedure calls to the server without the risk of exposing unwanted methods
// just because they satisfy the RPC interface
type Proxy struct {
	server Server
}

// New returns a new proxy
func New(server Server) *Proxy {
	return &Proxy{server: server}
}

// FocusGained should be called when a buffer gains focus
func (p *Proxy) FocusGained(event domain.Event, reply *string) error {
	return p.server.FocusGained(event, reply)
}

// OpenFile should be called when a file is opened
func (p *Proxy) OpenFile(event domain.Event, reply *string) error {
	return p.server.OpenFile(event, reply)
}

// SendHeartbeat can be called on writes, cursor moves, searches, etc. It
// informs the server that the coding session is still active
func (p *Proxy) SendHeartbeat(event domain.Event, reply *string) error {
	return p.server.SendHeartbeat(event, reply)
}

// EndSession should be called when the editor closes
func (p *Proxy) EndSession(event domain.Event, reply *string) error {
	return p.server.EndSession(event, reply)
}
