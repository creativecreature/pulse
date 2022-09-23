package session

import (
	"sync"
	"time"

	"code-harvest.conner.dev/internal/file"
)

type Session struct {
	mutex         sync.Mutex
	currentFile   *file.File
	lastHeartBeat int64
	openFiles     []*file.File
	StartedAt     int64                 `bson:"started_at"`
	EndedAt       int64                 `bson:"ended_at"`
	DurationMs    int64                 `bson:"duration_ms"`
	OS            string                `bson:"os"`
	Editor        string                `bson:"editor"`
	Files         map[string]*file.File `bson:"files"`
}

// Archives the currently opened file.
func (session *Session) archiveCurrentFile(time int64) {
	if session.currentFile == nil {
		return
	}

	session.currentFile.ClosedAt = time
	session.openFiles = append(session.openFiles, session.currentFile)
	session.currentFile = nil
}

// Aggregates all the files in the openFiles slice into the Files map
func (session *Session) aggregateFiles() {
	for _, f := range session.openFiles {
		currentFile, ok := session.Files[f.Path]
		if !ok {
			f.DurationMs = f.ClosedAt - f.OpenedAt
			session.Files[f.Path] = f
		} else {
			currentFile.DurationMs += f.ClosedAt - f.OpenedAt
		}
	}
}

func New(os, editor string) *Session {
	return &Session{
		lastHeartBeat: time.Now().UTC().UnixMilli(),
		StartedAt:     time.Now().UTC().UnixMilli(),
		OS:            os,
		Editor:        editor,
		Files:         make(map[string]*file.File),
	}
}

// UpdateCurrentFile sets the file that is currently being used.
func (session *Session) UpdateCurrentFile(file *file.File) {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	session.archiveCurrentFile(file.OpenedAt)
	session.currentFile = file
	session.lastHeartBeat = time.Now().UTC().UnixMilli()
}

// Heartbeat sets a timestamp for when the session was last active.
func (session *Session) Heartbeat() {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	session.lastHeartBeat = time.Now().UTC().UnixMilli()
}

// IsAlive returns true if the sessions last heartbeat plus the ttl is greater than time.Now().
func (session *Session) IsAlive(ttl int64) bool {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	return (session.lastHeartBeat + ttl) > time.Now().UTC().UnixMilli()
}

// End ends the coding session.
func (session *Session) End() {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	endedAt := time.Now().UTC().UnixMilli()
	session.archiveCurrentFile(endedAt)
	session.EndedAt = endedAt
	session.DurationMs = session.EndedAt - session.StartedAt

	if len(session.openFiles) < 1 {
		return
	}

	session.aggregateFiles()
}
