package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"code-harvest.conner.dev/internal/file"
	"code-harvest.conner.dev/internal/session"
	"code-harvest.conner.dev/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrWrongSession = errors.New("was called by a client that isn't considered active")
)

// Event represents the arguments that are passed to us by the client.
type Event struct {
	Id     string
	Path   string
	Editor string
	OS     string
}

type CodeHarvestApp struct {
	logger         *logger.Logger
	activeClientId string
	session        *session.Session
	mutex          sync.Mutex
	ctx            context.Context
	client         *mongo.Client
}

// Listens for SIGINT and SIGTERM signals. If we receive one of them we
// save the current session and inform the ECG and RPCServer to stop.
func (app *CodeHarvestApp) handleShutdown(errorChannel chan error) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Blocks until a signal is received.
	s := <-quit
	app.logger.PrintInfo("Preparing shutdown.", map[string]string{
		"signal": s.String(),
	})

	app.mutex.Lock()
	defer app.mutex.Unlock()

	// End the current session.
	if app.session != nil {
		app.session.End()
		app.saveSession()
	}

	// The ECG and RPCServer will be stopped when we publish to this channel.
	errorChannel <- nil
}

// Called by the ECG to determine whether the current session has gone stale or not.
func (app *CodeHarvestApp) checkHeartbeat() {
	app.logger.PrintDebug("Checking heartbeat", nil)
	if app.session != nil && !app.session.IsAlive(heartbeatTTL.Milliseconds()) {
		app.mutex.Lock()
		defer app.mutex.Unlock()
		app.session.End()
		app.saveSession()
	}
}

// Saves the current session and resets state
func (app *CodeHarvestApp) saveSession() {
	app.logger.PrintDebug("Saving the session.", nil)

	s := app.session
	app.activeClientId = ""
	app.session = nil

	if len(s.Files) < 1 {
		app.logger.PrintDebug("The session had no files.", nil)
		return
	}

	sessionCollection := app.client.Database("codeharvest").Collection("sessions")
	_, err := sessionCollection.InsertOne(context.Background(), s)
	if err != nil {
		app.logger.PrintError(err, nil)
	}

	app.logger.PrintDebug("The session was saved successfully.", nil)
}

// FocusGained should be called by the FocusGained autocommand. It gives us information
// about the currently active client. The duration of a coding session should not increase
// by the number of clients (VIM instances) we use. Only one will be tracked at a time.
// When we jump between them, it will switch the one we are counting time for.
func (app *CodeHarvestApp) FocusGained(event Event, reply *string) error {
	// The heartbeat timer could fire at the exact same time.
	app.mutex.Lock()
	defer app.mutex.Unlock()

	// When I jump between TMUX splits the *FocusGained* event in VIM will fire a
	// lot. I only want to end the current session, and create a new one, when I
	// open a new instance of VIM. If I'm, for :w
	//example, jumping between a VIM
	// split and a terminal with test output I don't want it to result in a new
	// coding session.
	if app.activeClientId == event.Id {
		app.logger.PrintDebug("Jumped back to the same instance of VIM.", nil)
		return nil
	}

	// This could be the first VIM instance which means there is no previous session.
	if app.session != nil {
		app.session.End()
		app.saveSession()
	}

	app.activeClientId = event.Id
	app.session = session.New(event.OS, event.Editor)

	// It could be an already existing VIM instance where a file buffer is already
	// open. If that is the case we can't count on getting the *OpenFile* event.
	// We might just be jumping between two VIM instances with one buffer each.
	f, err := file.New(event.Path)
	if err != nil {
		app.logger.PrintDebug("No file is currently being focused. Most likely a fresh VIM instance.", map[string]string{
			"path":  event.Path,
			"error": err.Error(),
		})
		return nil
	}

	app.session.UpdateCurrentFile(f)
	*reply = "Successfully updated the client being focused."
	return nil
}

// OpenFile should be called by the *BufEnter* autocommand.
func (app *CodeHarvestApp) OpenFile(event Event, reply *string) error {
	f, err := file.New(event.Path)
	if err != nil {
		app.logger.PrintDebug("Failed to create file from path.", map[string]string{
			"path":  event.Path,
			"error": err.Error(),
		})
		return nil
	}

	// To not collide with the heartbeat check that runs on an interval.
	app.mutex.Lock()
	defer app.mutex.Unlock()

	// The server won't receive any heartbeats if we open a buffer and then go AFK.
	// When that happens the session is ended. If we come back and either write the buffer,
	// or open a new file, we have to create a new session first.
	if app.session == nil {
		app.activeClientId = event.Id
		app.session = session.New(event.OS, event.Editor)
	}

	app.session.UpdateCurrentFile(f)
	*reply = "Successfully updated the current file."
	return nil
}

// SendHeartbeat should be called when we want to inform the server that the session
// is still active. If we, for example, only edit a single file for a long time we
// can send it on a *BufWrite* autocommand.
func (app *CodeHarvestApp) SendHeartbeat(event Event, reply *string) error {
	// In case the heartbeat check that runs on an interval occurs at the same time.
	app.mutex.Lock()
	defer app.mutex.Unlock()

	// If the session haven't expired (which happens if we AFK) we can just
	// update the last heartbeat.
	if app.session != nil {
		app.session.Heartbeat()
		*reply = "Successfully updated the current sessions heartbeat."
		return nil
	}

	// This scenario would occur if we write the buffer when we have been
	// inactive for more than 10 minutes. The server will have ended our coding
	// session. Therefore, we have to create a new one.
	app.activeClientId = event.Id
	app.session = session.New(event.OS, event.Editor)
	file, err := file.New(event.Path)
	if err != nil {
		app.logger.PrintDebug("Failed to create file from path.", map[string]string{
			"path":  event.Path,
			"error": err.Error(),
		})
		return nil
	}
	app.session.UpdateCurrentFile(file)

	*reply = "Heartbeat resulted in a new coding session."
	return nil
}

// EndSession should be called by the *VimLeave* autocommand to inform the server that the session is done.
func (app *CodeHarvestApp) EndSession(args struct{ Id string }, reply *string) error {
	app.mutex.Lock()
	defer app.mutex.Unlock()

	// We have reached an undesired state if we call end session and there is another
	// active client. It means that the events are sent in an incorrect order.
	if len(app.activeClientId) > 1 && app.activeClientId != args.Id {
		app.logger.PrintFatal(ErrWrongSession, map[string]string{
			"actualClientId":   app.activeClientId,
			"expectedClientId": args.Id,
		})
		return ErrWrongSession
	}

	// If we go AFK and don't send any heartbeats the session will have ended by
	// itself. This is different from the case above because if this happens there
	// shouldn't be another active client.
	if app.activeClientId == "" && app.session == nil {
		app.logger.PrintDebug("The session was already ended by the heartbeat check.", nil)
		return nil
	}

	// End the session by resetting the values.
	app.activeClientId = ""
	if app.session != nil {
		app.session.End()
		app.saveSession()
	}

	*reply = "The session was ended successfully."
	return nil
}
