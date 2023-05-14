package server_test

import (
	"errors"
	"io"
	"testing"

	"code-harvest.conner.dev/internal/domain"
	"code-harvest.conner.dev/internal/filereader"
	"code-harvest.conner.dev/internal/server"
	"code-harvest.conner.dev/pkg/clock"
	"code-harvest.conner.dev/pkg/logger"
)

func TestJumpingBetweenInstances(t *testing.T) {
	t.Parallel()

	mockStorage := &MockStorage{}
	mockMetadataReader := &MockFileReader{}

	a, err := server.New(
		"TestApp",
		server.WithLog(logger.New(io.Discard, logger.LevelOff)),
		server.WithMetadataReader(mockMetadataReader),
		server.WithStorage(mockStorage),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Open a new VIM instance
	reply := ""
	mockMetadataReader.file = mockFile{}
	a.FocusGained(domain.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Open a file in the first instance
	mockMetadataReader.file = mockFile{
		name:       "install.sh",
		filetype:   "bash",
		repository: "dotfiles",
	}
	a.OpenFile(domain.Event{
		Id:     "123",
		Path:   "/Users/conner/code/dotfiles/install.sh",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Open another vim instance in a new split. This should end the previous session.
	mockMetadataReader.file = mockFile{}
	a.FocusGained(domain.Event{
		Id:     "345",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Open a file in the second vim instance
	mockMetadataReader.file = mockFile{
		name:       "bootstrap.sh",
		filetype:   "bash",
		repository: "dotfiles",
	}
	a.OpenFile(domain.Event{
		Id:     "345",
		Path:   "/Users/conner/code/dotfiles/bootstrap.sh",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Move focus back to the first VIM instance. This should end the second session.
	mockMetadataReader.file = mockFile{
		name:       "install.sh",
		filetype:   "bash",
		repository: "dotfiles",
	}
	a.FocusGained(domain.Event{
		Id:     "123",
		Path:   "/Users/conner/code/dotfiles/install.sh",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// End the last session. We should now have 3 finished sessiona.
	a.EndSession(domain.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	expectedNumberOfSessions := 3
	storedSessions := mockStorage.Get()

	if len(storedSessions) != expectedNumberOfSessions {
		t.Errorf("expected len %d; got %d", expectedNumberOfSessions, len(storedSessions))
	}
}

func TestJumpBackAndForthToTheSameInstance(t *testing.T) {
	t.Parallel()

	mockStorage := &MockStorage{}
	mockMetadataReader := &MockFileReader{}

	a, err := server.New(
		"testApp",
		server.WithLog(logger.New(io.Discard, logger.LevelOff)),
		server.WithMetadataReader(mockMetadataReader),
		server.WithStorage(mockStorage),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Open a new instance of VIM
	reply := ""
	mockMetadataReader.file = mockFile{}
	a.FocusGained(domain.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Open a file
	mockMetadataReader.file = mockFile{
		name:       "install.sh",
		filetype:   "bash",
		repository: "dotfiles",
	}
	a.OpenFile(domain.Event{
		Id:     "123",
		Path:   "/Users/conner/code/dotfiles/install.sh",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Lets now imagine we opened another TMUX split to run testa. We then jump
	// back to VIM which will fire the focus gained event with the same client id.
	a.FocusGained(domain.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// We repeat the same thing again. Jump to another split in the terminal which makes
	// VIM lose focus and then back again - which will trigger another focus gained event.
	a.FocusGained(domain.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)
	mockMetadataReader.file = mockFile{
		name:       "bootstrap.sh",
		filetype:   "bash",
		repository: "dotfiles",
	}
	a.OpenFile(domain.Event{
		Id:     "123",
		Path:   "/Users/conner/code/dotfiles/bootstrap.sh",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Lets now end the session. This behaviour should *not* have resulted in any
	// new sessions being created. We only create a new session and end the current
	// one if we open VIM in a new split (to not count double time).
	a.EndSession(domain.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	expectedNumberOfSessions := 1
	storedSessions := mockStorage.Get()

	if len(storedSessions) != expectedNumberOfSessions {
		t.Errorf("expected len %d; got %d", expectedNumberOfSessions, len(storedSessions))
	}
}

func TestNoActivityShouldEndSession(t *testing.T) {
	t.Parallel()

	mockStorage := &MockStorage{}
	mockClock := &clock.MockClock{}
	mockMetadataReader := &MockFileReader{}
	mockMetadataReader.file = mockFile{}

	a, err := server.New(
		"testApp",
		server.WithLog(logger.New(io.Discard, logger.LevelOff)),
		server.WithClock(mockClock),
		server.WithMetadataReader(mockMetadataReader),
		server.WithStorage(mockStorage),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Send the initial focus event
	mockClock.SetTime(100)
	reply := ""
	a.FocusGained(domain.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	mockClock.SetTime(200)
	a.CheckHeartbeat()

	// Send an open file event. This should update the time for the last activity to 250.
	mockClock.SetTime(250)
	mockMetadataReader.file = mockFile{
		name:       "install.sh",
		filetype:   "bash",
		repository: "dotfiles",
	}
	a.OpenFile(domain.Event{
		Id:     "123",
		Path:   "/Users/conner/code/dotfiles/install.sh",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	// Perform another heartbeat check. Remember these checks does not update
	// the time for when we last saw activity in the session.
	mockClock.SetTime(300)
	a.CheckHeartbeat()

	// Heartbeat check that occurs 1 millisecond after the time of last activity
	// + ttl. This should result in the session being ended and saved.
	mockClock.SetTime(server.HeartbeatTTL.Milliseconds() + 250 + 1)
	a.CheckHeartbeat()

	mockClock.SetTime(server.HeartbeatTTL.Milliseconds() + 300)
	mockMetadataReader.file = mockFile{
		name:       "cleanup.sh",
		filetype:   "bash",
		repository: "dotfiles",
	}
	a.OpenFile(domain.Event{
		Id:     "123",
		Path:   "/Users/conner/code/dotfiles/cleanup.sh",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	mockClock.SetTime(server.HeartbeatTTL.Milliseconds() + 400)
	a.CheckHeartbeat()

	a.EndSession(domain.Event{
		Id:     "123",
		Path:   "",
		Editor: "nvim",
		OS:     "Linux",
	}, &reply)

	expectedNumberOfSessions := 2
	storedSessions := mockStorage.Get()

	if len(storedSessions) != expectedNumberOfSessions {
		t.Errorf("expected len %d; got %d", expectedNumberOfSessions, len(storedSessions))
	}
}

type MockStorage struct {
	sessions []*domain.Session
}

func (m *MockStorage) Connect() func() {
	return func() {}
}

func (m *MockStorage) Save(s interface{}) error {
	result, ok := s.(*domain.Session)
	if !ok {
		return errors.New("Failed to convert interface to slice of session pointers")
	}
	m.sessions = append(m.sessions, result)
	return nil
}

func (m *MockStorage) Get() []*domain.Session {
	return m.sessions
}

type mockFile struct {
	name       string
	filetype   string
	repository string
}

func (m mockFile) Name() string {
	return m.name
}

func (m mockFile) Filetype() string {
	return m.filetype
}

func (m mockFile) Repository() string {
	return m.repository
}

type MockFileReader struct {
	file filereader.File
}

func (f *MockFileReader) Read(path string) (filereader.File, error) {
	if f.file == nil {
		return mockFile{}, errors.New("metadata is nil")
	}
	return mockFile{}, nil
}
