package server

import "code-harvest.conner.dev/internal/shared"

// FocusGained should be called by the FocusGained autocommand. It gives us information
// about the currently active client. The duration of a coding session should not increase
// by the number of clients (VIM instances) we use. Only one will be tracked at a time.
func (app *App) FocusGained(event shared.Event, reply *string) error {
	// The heartbeat timer could fire at the exact same time.
	app.mutex.Lock()
	defer app.mutex.Unlock()

	app.lastHeartbeat = app.Clock.GetTime()

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
	app.createSession(event.OS, event.Editor)

	// It could be an already existing VIM instance where a file buffer is already
	// open. If that is the case we can't count on getting the *OpenFile* event.
	// We might just be jumping between two VIM instances with one buffer each.
	app.updateCurrentFile(event.Path)

	*reply = "Successfully updated the client being focused."
	return nil
}

// OpenFile should be called by the *BufEnter* autocommand.
func (app *App) OpenFile(event shared.Event, reply *string) error {
	app.log.PrintDebug("Received OpenFile event", map[string]string{
		"path": event.Path,
	})

	// To not collide with the heartbeat check that runs on an interval.
	app.mutex.Lock()
	defer app.mutex.Unlock()

	app.lastHeartbeat = app.Clock.GetTime()

	// The app won't receive any heartbeats if we open a buffer and then go AFK.
	// When that happens the session is ended. If we come back and either write the buffer,
	// or open a new file, we have to create a new session first.
	if app.session == nil {
		app.activeClientId = event.Id
		app.createSession(event.OS, event.Editor)
	}

	app.updateCurrentFile(event.Path)
	*reply = "Successfully updated the current file."
	return nil
}

// SendHeartbeat should be called when we want to inform the app that the session
// is still active. If we, for example, only edit a single file for a long time we
// can send it on a *BufWrite* autocommand.
func (app *App) SendHeartbeat(event shared.Event, reply *string) error {
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
		app.createSession(event.OS, event.Editor)
		app.updateCurrentFile(event.Path)
	}

	// Update the time for the last heartbeat.
	app.lastHeartbeat = app.Clock.GetTime()

	*reply = "Successfully sent heartbeat"
	return nil
}

// EndSession should be called by the *VimLeave* autocommand to inform the app that the session is done.
func (app *App) EndSession(event shared.Event, reply *string) error {
	app.mutex.Lock()
	defer app.mutex.Unlock()

	// We have reached an undesired state if we call end session and there is another
	// active client. It means that the events are sent in an incorrect order.
	if len(app.activeClientId) > 1 && app.activeClientId != event.Id {
		app.log.PrintFatal(ErrWrongSession, map[string]string{
			"actualClientId":   app.activeClientId,
			"expectedClientId": event.Id,
		})
		return ErrWrongSession
	}

	// If we go AFK and don't send any heartbeats the session will have ended by
	// itself. If we then come back and exit VIM we will get the EndSession event
	// but won't have any session that we are tracking time for.
	if app.activeClientId == "" && app.session == nil {
		message := "The session was already ended, or possibly never started. Was there a previous hearbeat check?"
		app.log.PrintDebug(message, nil)
		return nil
	}

	app.saveSession()

	*reply = "The session was ended successfully."
	return nil
}

// Called by the ECG to determine whether the current session has gone stale or not.
func (app *App) CheckHeartbeat() {
	app.log.PrintDebug("Checking heartbeat", nil)
	if app.session != nil && app.lastHeartbeat+HeartbeatTTL.Milliseconds() < app.Clock.GetTime() {
		app.mutex.Lock()
		defer app.mutex.Unlock()
		app.saveSession()
	}
}
