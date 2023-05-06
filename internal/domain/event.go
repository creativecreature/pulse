package domain

// Event that is sent from the client to the server
type Event struct {
	Id     string
	Path   string
	Editor string
	OS     string
}
