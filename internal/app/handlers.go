package app

import (
	"errors"

	"code-harvest.conner.dev/internal/models"
	"code-harvest.conner.dev/internal/shared"
)

func (app *app) startNewSession(os, editor string) {
	app.session = models.NewSession(app.clock.GetTime(), os, editor)
}

func (app *app) updateCurrentFile(path string) {
	openedAt := app.clock.GetTime()

	fileMetadata, err := app.metadataReader.Read(path)
	if err != nil {
		app.log.PrintDebug("Could not extract metadata for the path", map[string]string{
			"reason": err.Error(),
		})
		return
	}

	file := models.NewFile(
		fileMetadata.Filename,
		fileMetadata.RepositoryName,
		fileMetadata.Filetype,
		path,
		openedAt,
	)

	// Update the current file.
	if currentFile := app.session.FileStack.Peek(); currentFile != nil {
		currentFile.ClosedAt = openedAt
	}
	app.session.FileStack.Push(file)
	app.log.PrintDebug("Successfully updated the current file", map[string]string{
		"path": path,
	})
}

func (app *app) saveSession() {
	// Regardless of how we exit this function we want to reset these values.
	defer func() {
		app.activeClientId = ""
		app.session = nil
	}()

	if app.session == nil {
		app.log.PrintDebug("There was no session to save.", nil)
		return
	}

	app.log.PrintDebug("Saving the session.", nil)

	// Set session duration and set closed at for the current file.
	endedAt := app.clock.GetTime()
	if currentFile := app.session.FileStack.Peek(); currentFile != nil {
		currentFile.ClosedAt = endedAt
	}
	app.session.EndedAt = endedAt
	app.session.DurationMs = app.session.EndedAt - app.session.StartedAt

	// Whenever we open new a buffer that have a corresponding file on disk we
	// push it to the sessions file stack. Each file can appear more than once.
	// Before we save the session we aggregate all the edits of the same file
	// into a map with a total duration of the time we've spent in that file.
	for app.session.FileStack.Len() > 0 {
		file := app.session.FileStack.Pop()
		aggregatedFile, exists := app.session.AggregatedFiles[file.Path]
		if !exists {
			file.DurationMs = file.ClosedAt - file.OpenedAt
			app.session.AggregatedFiles[file.Path] = file
		} else {
			aggregatedFile.DurationMs += file.ClosedAt - file.OpenedAt
		}
	}

	if len(app.session.AggregatedFiles) < 1 {
		app.log.PrintDebug("The session had no files.", map[string]string{
			"clientId": app.activeClientId,
		})
		return
	}

	err := app.storage.Save(app.session)
	if err != nil {
		app.log.PrintError(err, nil)
	}
}

// FocusGained should be called by the FocusGained autocommand. It gives us information
// about the currently active client. The duration of a coding session should not increase
// by the number of clients (VIM instances) we use. Only one will be tracked at a time.
func (app *app) FocusGained(event shared.Event, reply *string) error {
	// The heartbeat timer could fire at the exact same time.
	app.mutex.Lock()
	defer app.mutex.Unlock()

	app.lastHeartbeat = app.clock.GetTime()

	// When I jump between TMUX splits the *FocusGained* event in VIM will fire a
	// lot. I only want to end the current session, and create a new one, when I
	// open a new instance of VIM. If I'm, for example, jumping between a VIM split
	// and a terminal with test output I don't want it to result in a new coding session.
	if app.activeClientId == event.Id {
		app.log.PrintDebug("Jumped back to the same instance of VIM.", nil)
		return nil
	}

	// If the focus event is for the first instance of VIM we won't have any previous session.
	// That only occurs when using multiple splits with multiple instances of VIM.
	if app.session != nil {
		app.saveSession()
	}

	app.activeClientId = event.Id
	app.startNewSession(event.OS, event.Editor)

	// It could be an already existing VIM instance where a file buffer is already
	// open. If that is the case we can't count on getting the *OpenFile* event.
	// We might just be jumping between two VIM instances with one buffer each.
	app.updateCurrentFile(event.Path)

	*reply = "Successfully updated the client being focused."
	return nil
}

// OpenFile should be called by the *BufEnter* autocommand.
func (app *app) OpenFile(event shared.Event, reply *string) error {
	app.log.PrintDebug("Received OpenFile event", map[string]string{
		"path": event.Path,
	})

	// To not collide with the heartbeat check that runs on an interval.
	app.mutex.Lock()
	defer app.mutex.Unlock()

	app.lastHeartbeat = app.clock.GetTime()

	// The app won't receive any heartbeats if we open a buffer and then go AFK.
	// When that happens the session is ended. If we come back and either write the buffer,
	// or open a new file, we have to create a new session first.
	if app.session == nil {
		app.activeClientId = event.Id
		app.startNewSession(event.OS, event.Editor)
	}

	app.updateCurrentFile(event.Path)
	*reply = "Successfully updated the current file."
	return nil
}

// SendHeartbeat should be called when we want to inform the app that the session
// is still active. If we, for example, only edit a single file for a long time we
// can send it on a *BufWrite* autocommand.
func (app *app) SendHeartbeat(event shared.Event, reply *string) error {
	// In case the heartbeat check that runs on an interval occurs at the same time.
	app.mutex.Lock()
	defer app.mutex.Unlock()

	// This scenario would occur if we write the buffer when we have been
	// inactive for more than 10 minutes. The app will have ended our coding
	// session. Therefore, we have to create a new one.
	if app.session == nil {
		message := "The session was ended by a previous heartbeat check. Creating a new one."
		app.log.PrintDebug(message, map[string]string{
			"clientId": event.Id,
			"path":     event.Path,
		})
		app.activeClientId = event.Id
		app.startNewSession(event.OS, event.Editor)
		app.updateCurrentFile(event.Path)
	}

	// Update the time for the last heartbeat.
	app.lastHeartbeat = app.clock.GetTime()

	*reply = "Successfully sent heartbeat"
	return nil
}

// EndSession should be called by the *VimLeave* autocommand to inform the app that the session is done.
func (app *app) EndSession(event shared.Event, reply *string) error {
	app.mutex.Lock()
	defer app.mutex.Unlock()

	// We have reached an undesired state if we call end session and there is another
	// active client. It means that the events are sent in an incorrect order.
	if len(app.activeClientId) > 1 && app.activeClientId != event.Id {
		app.log.PrintFatal(errors.New("was called by a client that isn't considered active"), map[string]string{
			"actualClientId":   app.activeClientId,
			"expectedClientId": event.Id,
		})
	}

	// If we go AFK and don't send any heartbeats the session will have ended by
	// itself. If we then come back and exit VIM we will get the EndSession event
	// but won't have any session that we are tracking time for.
	if app.activeClientId == "" && app.session == nil {
		message := "The session was already ended, or possibly never started. Was there a previous heatbeat check?"
		app.log.PrintDebug(message, nil)
		return nil
	}

	app.saveSession()

	*reply = "The session was ended successfully."
	return nil
}
