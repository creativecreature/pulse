package shared

type Server interface {
	FocusGained(event Event, reply *string) error
	OpenFile(event Event, reply *string) error
	SendHeartbeat(event Event, reply *string) error
	EndSession(event Event, reply *string) error
}

// The proxy is used to expose certain methods on the server.
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
