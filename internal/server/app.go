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
	"code-harvest.conner.dev/pkg/clock"
	"code-harvest.conner.dev/pkg/logger"
)

var HeartbeatTTL = time.Minute * 10
var heartbeatInterval = time.Second * 10

var (
	ErrWrongSession = errors.New("was called by a client that isn't considered active")
)

type App struct {
	Clock          clock.Clock
	MetadataReader MetadataReader
	mutex          sync.Mutex
	activeClientId string
	lastHeartbeat  int64
	session        *models.Session
	log            *logger.Logger
	storage        Storage
}

func (app *App) archiveCurrentFile(closedAt int64) {
	if app.session.CurrentFile != nil {
		app.session.CurrentFile.ClosedAt = closedAt
		app.session.OpenFiles = append(app.session.OpenFiles, app.session.CurrentFile)
	}
}

func (app *App) updateCurrentFile(path string) {
	openedAt := app.Clock.GetTime()

	fileMetadata, err := app.MetadataReader.Read(path)
	if err != nil {
		app.log.PrintDebug("Could not extract metadata for the path", map[string]string{
			"reason": err.Error(),
		})
		return
	}

	file := models.File{
		Name:       fileMetadata.Filename,
		Repository: fileMetadata.RepositoryName,
		Filetype:   fileMetadata.Filetype,
		Path:       path,
		OpenedAt:   openedAt,
		ClosedAt:   0,
	}

	// Update the current file.
	app.archiveCurrentFile(openedAt)
	app.session.CurrentFile = &file
	app.log.PrintDebug("Successfully updated the current file", map[string]string{
		"path": path,
	})
}

func (app *App) createSession(os, editor string) {
	app.session = &models.Session{
		StartedAt: time.Now().UTC().UnixMilli(),
		OS:        os,
		Editor:    editor,
		Files:     make(map[string]*models.File),
	}
}

func (app *App) saveSession() {
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

	// Set session duration and archive the current file.
	endedAt := app.Clock.GetTime()
	app.archiveCurrentFile(endedAt)
	app.session.EndedAt = endedAt
	app.session.DurationMs = app.session.EndedAt - app.session.StartedAt

	// The OpenFiles list reflects all files we've opened. Each file has a
	// OpenedAt and ClosedAt property. Every file can appear more than once.
	// Before we save the session we aggregate this into a map where the key
	// is the name of the file and the value is a File with a merged duration
	// for all edits.
	if len(app.session.OpenFiles) > 0 {
		for _, f := range app.session.OpenFiles {
			currentFile, ok := app.session.Files[f.Path]
			if !ok {
				f.DurationMs = f.ClosedAt - f.OpenedAt
				app.session.Files[f.Path] = f
			} else {
				currentFile.DurationMs += f.ClosedAt - f.OpenedAt
			}
		}
	}

	if len(app.session.Files) < 1 {
		app.log.PrintDebug("The session had no files.", map[string]string{
			"clientId": app.activeClientId,
		})
		fmt.Println(app.session.Files)
		return
	}

	err := app.storage.Save(app.session)
	if err != nil {
		app.log.PrintError(err, nil)
	}
}

func New(log *logger.Logger, storage Storage) *App {
	return &App{
		log:            log,
		storage:        storage,
		Clock:          clock.New(),
		MetadataReader: FileMetadataReader{},
	}
}

func (app *App) Start(port string) error {
	proxy := shared.NewServerProxy(app)
	err := rpc.RegisterName(shared.ServerName, proxy)
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
			app.CheckHeartbeat()
		case <-quit:
			run = false
		}
	}

	ecg.Stop()
	return listener.Close()
}
