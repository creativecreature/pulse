package server

import (
	"github.com/creativecreature/pulse"
)

func (s *Server) openFile(event pulse.Event) {
	gitFile, gitFileErr := s.fileReader.GitFile(event.Path)
	if gitFileErr != nil {
		return
	}

	if s.activeBuffer != nil {
		if s.activeBuffer.Filepath == gitFile.Path && s.activeBuffer.Repository == gitFile.Repository {
			s.log.Debug("This buffer is already considered active.",
				"path", gitFile.Path,
				"repository", gitFile.Repository,
				"editor_id", event.EditorID,
				"editor", event.Editor,
				"os", event.OS,
			)
			return
		}
	}

	s.saveBuffer()
	buf := pulse.NewBuffer(
		gitFile.Name,
		gitFile.Repository,
		gitFile.Filetype,
		gitFile.Path,
		s.clock.Now(),
	)
	s.activeBuffer = &buf
}

// FocusGained is invoked by the FocusGained autocommand.
func (s *Server) FocusGained(event pulse.Event, reply *string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.lastHeartbeat = s.clock.Now()
	s.log.Debug("Received FocusGained event.",
		"editor_id", event.EditorID,
		"editor", event.Editor,
		"os", event.OS,
	)

	if event.Path == "" {
		return
	}

	s.openFile(event)
	*reply = "Successfully updated the client being focused."
}

// OpenFile gets invoked by the *BufEnter* autocommand.
func (s *Server) OpenFile(event pulse.Event, reply *string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.lastHeartbeat = s.clock.Now()
	s.log.Debug("Received OpenFile event.",
		"editor_id", event.EditorID,
		"editor", event.Editor,
		"os", event.OS,
	)

	if event.Path == "" {
		return
	}

	s.openFile(event)
	*reply = "Successfully updated the current file."
}

// SendHeartbeat can be called for events such as buffer writes and cursor moves.
// Its purpose is to notify the server that the current session remains active.
// The server ends the session if it doesn't receive a heartbeat for 10 minutes.
func (s *Server) SendHeartbeat(event pulse.Event, reply *string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.lastHeartbeat = s.clock.Now()
	s.log.Debug("Received heartbeat.",
		"editor_id", event.EditorID,
		"editor", event.Editor,
		"os", event.OS,
	)
	*reply = "Successfully sent heartbeat"
}

// EndSession should be called by the *VimLeave* autocommand.
func (s *Server) EndSession(event pulse.Event, reply *string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.log.Debug("Received EndSession event",
		"editor_id", event.EditorID,
		"editor", event.Editor,
		"os", event.OS,
	)
	s.saveBuffer()
	*reply = "The session was ended successfully."
}
