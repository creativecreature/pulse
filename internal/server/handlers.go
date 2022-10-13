package server

import "code-harvest.conner.dev/internal/shared"

// Event represents the arguments that are passed to us by the client.

// The Handlers struct is essentially a proxy for the functions
// that I want to expose to remote procedure calls.
type Handlers struct {
	app *App
}

func NewHandlers(app *App) *Handlers {
	return &Handlers{app: app}
}

func (h *Handlers) FocusGained(event shared.Event, reply *string) error {
	return h.app.FocusGained(event, reply)
}

func (h *Handlers) OpenFile(event shared.Event, reply *string) error {
	return h.app.OpenFile(event, reply)
}

func (h *Handlers) SendHeartbeat(event shared.Event, reply *string) error {
	return h.app.SendHeartbeat(event, reply)
}

func (h *Handlers) EndSession(args struct{ Id string }, reply *string) error {
	return h.app.EndSession(args, reply)
}
