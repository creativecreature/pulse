package domain

// Event represents the events we receive from the clients.
type Event struct {
	Id     string
	Path   string
	Editor string
	OS     string
}
