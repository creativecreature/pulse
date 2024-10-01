package pulse

// Event represents the events we receive from the editor.
type Event struct {
	EditorID string
	Path     string
	Filetype string
	Editor   string
	OS       string
}
