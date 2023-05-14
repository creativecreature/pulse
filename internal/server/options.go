package server

import (
	"errors"

	"code-harvest.conner.dev/internal/storage"
	"code-harvest.conner.dev/pkg/clock"
)

type option func(*server) error

func WithClock(clock clock.Clock) option {
	return func(a *server) error {
		if clock == nil {
			return errors.New("clock is nil")
		}
		a.clock = clock
		return nil
	}
}

func WithMetadataReader(reader FileMetadataReader) option {
	return func(a *server) error {
		if reader == nil {
			return errors.New("reader is nil")
		}
		a.metadataReader = reader
		return nil
	}
}

func WithStorage(storage storage.Storage) option {
	return func(a *server) error {
		if storage == nil {
			return errors.New("storage is nil")
		}
		a.storage = storage
		return nil
	}
}

type Log interface {
	PrintDebug(message string, properties map[string]string)
	PrintInfo(message string, properties map[string]string)
	PrintError(err error, properties map[string]string)
	PrintFatal(err error, properties map[string]string)
}

func WithLog(log Log) option {
	return func(a *server) error {
		if log == nil {
			return errors.New("log is nil")
		}
		a.log = log
		return nil
	}
}

func New(serverName string, opts ...option) (*server, error) {
	a := &server{
		serverName:     serverName,
		clock:          clock.New(),
		metadataReader: newFileReader(filesystem{}),
	}
	for _, opt := range opts {
		err := opt(a)
		if err != nil {
			return &server{}, err
		}
	}
	return a, nil
}
