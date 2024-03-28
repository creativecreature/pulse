package server

import (
	"errors"
	"strconv"

	codeharvest "github.com/creativecreature/code-harvest"
)

// FocusGained is invoked by the FocusGained autocommand. It gives
// us information about the currently active client. The duration
// of a coding session should not increase by the number of clients
// (neovim instances). Only one will be tracked at a time.
func (s *Server) FocusGained(event codeharvest.Event, reply *string) {
	// Lock the mutex to prevent race conditions with the heartbeat check.
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.lastHeartbeat = s.clock.GetTime()

	// The FocusGained event will be triggered when I switch back to an active
	// editor from another TMUX split. However, the intent is to only terminate
	// the current session, and initiate a new one, if I'm opening another neovim
	// instance. If the FocusGained event is firing because I'm jumping back and
	// forth between a tmux split with test output I don't want it to result in
	// the creation of several new coding sessions.
	if s.activeEditor == event.EditorID {
		s.log.PrintDebug("Jumped back to the same neovim instance", nil)
		return
	}

	// If we already have an existing session active, we wont create
	// a new one until it actually opens a buffer with a path.
	gitFile, gitFileErr := s.fileReader.GitFile(event.Path)
	if s.activeEditor != "" && gitFileErr != nil {
		return
	}

	// Check to see if we have another instance of neovim that is
	// running in another tmux pane. If so, we'll stop recording
	// time for that session before creating a new one.
	if s.activeEditor != "" {
		// Pause the current session if we have a valid path.
		s.activeSessions[s.activeEditor].Pause(s.clock.GetTime())
	}
	s.activeEditor = event.EditorID

	// Check if we've paused this session. In that case, we'll resume it.
	if session, ok := s.activeSessions[event.EditorID]; ok {
		s.log.PrintDebug("Resuming session.", nil)
		session.Resume(s.clock.GetTime())
		return
	}

	s.startNewSession(event.EditorID, event.OS, event.Editor)
	// It could be an already existing neovim instance where a file buffer is already
	// open. If that is the case we can't count on getting the *OpenFile* event.
	// We might just be jumping between two neovim instances with one buffer each.
	if gitFileErr != nil {
		return
	}
	s.setActiveBuffer(gitFile)
	*reply = "Successfully updated the client being focused."
}

// OpenFile gets invoked by the *BufEnter* autocommand.
func (s *Server) OpenFile(event codeharvest.Event, reply *string) {
	// The FocusGained autocommand wont fire in some terminals,
	// or if focus-events aren't enabled in TMUX.
	s.FocusGained(event, reply)
	s.log.PrintDebug("Received OpenFile event", map[string]string{
		"path":   event.Path,
		"editor": event.EditorID,
	})

	if event.Path == "" {
		return
	}

	// Lock the mutex to prevent race conditions with the heartbeat check.
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// If a new file was opened it means that the session is still active.
	s.lastHeartbeat = s.clock.GetTime()
	gitFile, gitFileErr := s.fileReader.GitFile(event.Path)
	if gitFileErr != nil {
		return
	}
	s.setActiveBuffer(gitFile)
	*reply = "Successfully updated the current file."
}

// SendHeartbeat can be called for events such as buffer writes and cursor moves.
// Its purpose is to notify the server that the current session remains active.
// The server ends the session if it doesn't receive a heartbeat for 10 minutes.
func (s *Server) SendHeartbeat(event codeharvest.Event, reply *string) {
	// Lock the mutex to prevent race conditions with the heartbeat check.
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// This is to handle the case where the server would have ended the clients
	// session due to inactivity. When a session ends it is written to disk and
	// can't be resumed. Therefore, we'll have to create a new coding session.
	if s.activeEditor == "" {
		// We'll wait for an actual file to be opened before we create another session.
		gitFile, gitFileErr := s.fileReader.GitFile(event.Path)
		if gitFileErr != nil {
			return
		}
		s.log.PrintDebug("The session was ended by a heartbeat check. Creating a new one.", map[string]string{
			"editorID": event.EditorID,
			"path":     event.Path,
		})
		s.activeEditor = event.EditorID
		s.startNewSession(event.EditorID, event.OS, event.Editor)
		s.setActiveBuffer(gitFile)
	}

	// Update the time for the last heartbeat.
	s.lastHeartbeat = s.clock.GetTime()
	*reply = "Successfully sent heartbeat"
}

// EndSession should be called by the *VimLeave* autocommand.
func (s *Server) EndSession(event codeharvest.Event, reply *string) {
	s.log.PrintDebug("Received EndSession event", map[string]string{
		"editor": event.EditorID,
	})

	// Lock the mutex to prevent race conditions with the heartbeat check.
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.activeEditor != "" && s.activeEditor != event.EditorID {
		s.log.PrintFatal(errors.New("was called by a client that isn't considered active"), map[string]string{
			"actualClientId":   s.activeEditor,
			"expectedClientId": event.EditorID,
		})
	}

	// This could be the first event after more than ten minutes of inactivity.
	// If that is the case, the server will have ended the session already.
	if s.activeEditor == "" {
		message := "The session was already ended by the server."
		s.log.PrintDebug(message, nil)
		return
	}

	s.saveActiveSession()
	*reply = "The session was ended successfully."
}

// CheckHeartbeat is used to check if the session has been inactive for more than
// ten minutes. If that is the case, the session will be terminated and saved to disk.
func (s *Server) CheckHeartbeat() {
	s.log.PrintDebug("Checking heartbeat", nil)
	if s.activeEditor == "" {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.lastHeartbeat+HeartbeatTTL.Milliseconds() < s.clock.GetTime() {
		s.log.PrintInfo("Ending all active sessions due to inactivity", map[string]string{
			"last_heartbeat": strconv.FormatInt(s.lastHeartbeat, 10),
			"current_time":   strconv.FormatInt(s.clock.GetTime(), 10),
		})
		s.saveAllSessions()
		s.activeEditor = ""
	}
}
