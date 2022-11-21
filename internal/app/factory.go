package app

import (
	"errors"

	"code-harvest.conner.dev/pkg/clock"
	"code-harvest.conner.dev/pkg/logger"
)

type option func(*app) error

func WithClock(clock clock.Clock) option {
	return func(a *app) error {
		if clock == nil {
			return errors.New("clock is nil")
		}
		a.clock = clock
		return nil
	}
}

func WithMetadataReader(reader FileMetadataReader) option {
	return func(a *app) error {
		if reader == nil {
			return errors.New("reader is nil")
		}
		a.metadataReader = reader
		return nil
	}
}

func WithStorage(storage storage) option {
	return func(a *app) error {
		if storage == nil {
			return errors.New("storage is nil")
		}
		a.storage = storage
		return nil
	}
}

func WithLog(log *logger.Logger) option {
	return func(a *app) error {
		if log == nil {
			return errors.New("log is nil")
		}
		a.log = log
		return nil
	}
}

func New(opts ...option) (*app, error) {
	a := &app{
		clock:          clock.New(),
		metadataReader: newFileReader(filesystem{}),
	}
	for _, opt := range opts {
		err := opt(a)
		if err != nil {
			return &app{}, err
		}
	}
	return a, nil
}
