package server

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"code-harvest.conner.dev/internal/models"
	"code-harvest.conner.dev/internal/shared"
	"code-harvest.conner.dev/internal/storage"
	"code-harvest.conner.dev/pkg/logger"
)

var heartbeatTTL = time.Minute * 10
var heartbeatInterval = time.Second * 10

var (
	ErrWrongSession = errors.New("was called by a client that isn't considered active")
)

type App struct {
	activeClientId string
	session        *models.Session
	mutex          sync.Mutex
	log            *logger.Logger
	storage        storage.Storage
}

// Called by the ECG to determine whether the current session has gone stale or not.
func (app *App) checkHeartbeat() {
	app.log.PrintDebug("Checking heartbeat", nil)
	if app.session != nil && !app.session.IsAlive(heartbeatTTL.Milliseconds()) {
		app.mutex.Lock()
		defer app.mutex.Unlock()
		app.session.End()
		app.saveSession()
	}
}

// Saves the current session and resets state
func (app *App) saveSession() {
	app.log.PrintDebug("Saving the session.", nil)

	s := app.session
	app.activeClientId = ""
	app.session = nil

	if len(s.Files) < 1 {
		app.log.PrintDebug("The session had no files.", nil)
		return
	}

	err := app.storage.Save(s)
	if err != nil {
		app.log.PrintError(err, nil)
	}

	app.log.PrintDebug("The session was saved successfully.", nil)
}

// FocusGained should be called by the FocusGained autocommand. It gives us information
// about the currently active client. The duration of a coding session should not increase
// by the number of clients (VIM instances) we use. Only one will be tracked at a time.
// When we jump between them, it will switch the one we are counting time for.
func (app *App) FocusGained(event shared.Event, reply *string) error {
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
		app.log.PrintDebug("Jumped back to the same instance of VIM.", nil)
		return nil
	}

	// This could be the first VIM instance which means there is no previous session.
	if app.session != nil {
		app.session.End()
		app.saveSession()
	}

	app.activeClientId = event.Id
	app.session = models.NewSession(event.OS, event.Editor)

	// It could be an already existing VIM instance where a file buffer is already
	// open. If that is the case we can't count on getting the *OpenFile* event.
	// We might just be jumping between two VIM instances with one buffer each.
	f, err := models.NewFile(event.Path)
	if err != nil {
		app.log.PrintDebug("No file is currently being focused. Most likely a fresh VIM instance.", map[string]string{
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
func (app *App) OpenFile(event shared.Event, reply *string) error {
	f, err := models.NewFile(event.Path)
	if err != nil {
		app.log.PrintDebug("Failed to create file from path.", map[string]string{
			"path":  event.Path,
			"error": err.Error(),
		})
		return nil
	}

	// To not collide with the heartbeat check that runs on an interval.
	app.mutex.Lock()
	defer app.mutex.Unlock()

	// The app won't receive any heartbeats if we open a buffer and then go AFK.
	// When that hserverens the session is ended. If we come back and either write the buffer,
	// or open a new file, we have to create a new session first.
	if app.session == nil {
		app.activeClientId = event.Id
		app.session = models.NewSession(event.OS, event.Editor)
	}

	app.session.UpdateCurrentFile(f)
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

	// If the session haven't expired (which hserverens if we AFK) we can just
	// update the last heartbeat.
	if app.session != nil {
		app.session.Heartbeat()
		*reply = "Successfully updated the current sessions heartbeat."
		return nil
	}

	// This scenario would occur if we write the buffer when we have been
	// inactive for more than 10 minutes. The app will have ended our coding
	// session. Therefore, we have to create a new one.
	app.activeClientId = event.Id
	app.session = models.NewSession(event.OS, event.Editor)
	file, err := models.NewFile(event.Path)
	if err != nil {
		app.log.PrintDebug("Failed to create file from path.", map[string]string{
			"path":  event.Path,
			"error": err.Error(),
		})
		return nil
	}
	app.session.UpdateCurrentFile(file)

	*reply = "Heartbeat resulted in a new coding session."
	return nil
}

// EndSession should be called by the *VimLeave* autocommand to inform the app that the session is done.
func (app *App) EndSession(args struct{ Id string }, reply *string) error {
	app.mutex.Lock()
	defer app.mutex.Unlock()

	// We have reached an undesired state if we call end session and there is another
	// active client. It means that the events are sent in an incorrect order.
	if len(app.activeClientId) > 1 && app.activeClientId != args.Id {
		app.log.PrintFatal(ErrWrongSession, map[string]string{
			"actualClientId":   app.activeClientId,
			"expectedClientId": args.Id,
		})
		return ErrWrongSession
	}

	// If we go AFK and don't send any heartbeats the session will have ended by
	// itself. This is different from the case above because if this hserverens there
	// shouldn't be another active client.
	if app.activeClientId == "" && app.session == nil {
		message := "The session was already ended, or possibly never started. Was there a previous hearbeat check?"
		app.log.PrintDebug(message, nil)
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

func New(log *logger.Logger, storage storage.Storage) *App {
	return &App{log: log, storage: storage}
}

func (app *App) Start(port string) error {
	handlers := NewHandlers(app)
	err := rpc.RegisterName(shared.ServerName, handlers)
	if err != nil {
		return err
	}

	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return err
	}

	http.Serve(listener, nil)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	ecg := time.NewTicker(heartbeatInterval)

	run := true
	for run {
		select {
		case <-ecg.C:
			app.checkHeartbeat()
		case <-quit:
			run = false
		}
	}

	ecg.Stop()
	return listener.Close()
}
