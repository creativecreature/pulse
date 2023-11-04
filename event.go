package codeharvest

// Event represents the events we receive from the clients.
type Event struct {
	ID     string
	Path   string
	Editor string
	OS     string
}
