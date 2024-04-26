package server

import (
	"github.com/creativecreature/pulse"
)

// FocusGained is invoked by the FocusGained autocommand. It is used to set the
// editor that is currently active. Even though we might have multiple editors
// open at any given time, we'll only count time for one.
func (s *Server) FocusGained(event pulse.Event, reply *string) {
	s.log.Debug("Received FocusGained event.",
		"editor_id", event.EditorID,
		"editor", event.Editor,
		"os", event.OS,
	)

	s.checkHeartbeat()
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.lastHeartbeat = s.clock.Now()

	// The FocusGained event will be triggered when I switch back to an active
	// editor from another TMUX split. However, the intent is to only terminate
	// the current session, and initiate a new one, if I'm opening another neovim
	// instance. If the FocusGained event is firing because I'm jumping back and
	// forth between a tmux split with test output I don't want it to result in
	// the creation of several new coding sessions.
	if s.activeEditorID == event.EditorID {
		s.log.Debug("Jumped back to the same editor process.",
			"editor_id", event.EditorID,
			"editor", event.Editor,
			"os", event.OS,
		)
		return
	}

	// If we already have an existing session that we're counting time for, we wont
	// create a new one until it actually opens a buffer that is backed by a file.
	_, gitFileErr := s.fileReader.GitFile(event.Path)
	if s.activeEditorID != "" && gitFileErr != nil {
		s.log.Debug("Waiting for a file-backed buffer to be opened.",
			"editor_id", event.EditorID,
			"editor", event.Editor,
			"os", event.OS,
		)
		return
	}

	// Check to see if we have another instance of neovim that is running in a different tmux
	// pane. If so, we'll stop recording time for that session before creating a new one.
	if s.activeEditorID != "" {
		s.activeSessions[s.activeEditorID].Pause(s.clock.Now())
		s.log.Debug("Pausing session.",
			"editor_id", s.activeEditorID,
			"editor", s.activeSessions[s.activeEditorID].Editor,
			"os", s.activeSessions[s.activeEditorID].OS,
		)
	}

	// Check if this is an already existing session that we've paused. If that is the case, we'll resume it.
	s.activeEditorID = event.EditorID
	if session, ok := s.activeSessions[event.EditorID]; ok {
		s.log.Debug("Resuming session.",
			"editor_id", event.EditorID,
			"editor", event.Editor,
			"os", event.OS,
		)
		session.Resume(s.clock.Now())
		return
	}

	s.createSession(event.EditorID, event.OS, event.Editor)
	*reply = "Successfully updated the client being focused."
}

// OpenFile gets invoked by the *BufEnter* autocommand.
func (s *Server) OpenFile(event pulse.Event, reply *string) {
	if event.Path == "" {
		return
	}

	s.checkHeartbeat()
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.lastHeartbeat = s.clock.Now()

	// The editor could have been inactive, while focused, for 10 minutes.
	// That would end the session, and we could get a OpenFile event without
	// a preceding FocusGained.
	if s.activeEditorID != event.EditorID {
		s.mutex.Unlock()
		s.FocusGained(event, reply)
		s.mutex.Lock()
	}

	// If a new file was opened it means that the session is still active.
	gitFile, gitFileErr := s.fileReader.GitFile(event.Path)
	if gitFileErr != nil {
		s.log.Debug("Failed to get git file.",
			"path", event.Path,
			"err", gitFileErr.Error(),
			"editor_id", event.EditorID,
			"editor", event.Editor,
			"os", event.OS,
		)
		return
	}
	s.setActiveBuffer(gitFile)
	*reply = "Successfully updated the current file."
}

// SendHeartbeat can be called for events such as buffer writes and cursor moves.
// Its purpose is to notify the server that the current session remains active.
// The server ends the session if it doesn't receive a heartbeat for 10 minutes.
func (s *Server) SendHeartbeat(event pulse.Event, reply *string) {
	s.log.Debug("Received heartbeat.",
		"editor_id", event.EditorID,
		"editor", event.Editor,
		"os", event.OS,
	)

	s.checkHeartbeat()
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.lastHeartbeat = s.clock.Now()

	// This is to handle the case where the server would have ended the clients
	// session due to inactivity. When a session ends it is written to disk and
	// can't be resumed. Therefore, we'll have to create a new coding session.
	if s.activeEditorID == "" {
		// We'll wait for an actual file to be opened before we create another session.
		gitFile, gitFileErr := s.fileReader.GitFile(event.Path)
		if gitFileErr != nil {
			return
		}
		s.log.Debug(
			"The session was ended by a heartbeat check. Creating a new one.",
			"path", event.Path,
			"editor_id", event.EditorID,
			"editor", event.Editor,
			"os", event.OS,
		)
		s.activeEditorID = event.EditorID
		s.createSession(event.EditorID, event.OS, event.Editor)
		s.setActiveBuffer(gitFile)
	}

	*reply = "Successfully sent heartbeat"
}

// EndSession should be called by the *VimLeave* autocommand.
func (s *Server) EndSession(event pulse.Event, reply *string) {
	s.log.Debug("Received EndSession event",
		"editor_id", event.EditorID,
		"editor", event.Editor,
		"os", event.OS,
	)

	s.checkHeartbeat()
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.activeEditorID != "" && s.activeEditorID != event.EditorID {
		s.log.Debug("EndSession was called by an editor that wasn't considered active.",
			"active_editor_id", s.activeEditorID,
			"editor_id", event.EditorID,
			"editor", event.Editor,
			"os", event.OS,
		)
		return
	}

	// This could be the first event after more than ten minutes of inactivity.
	// If that is the case, the server will have ended the session already.
	if s.activeEditorID == "" {
		s.log.Debug("The session was already ended by the server.",
			"editor_id", event.EditorID,
			"editor", event.Editor,
			"os", event.OS,
		)
		return
	}

	s.saveActiveSession()
	*reply = "The session was ended successfully."
}
