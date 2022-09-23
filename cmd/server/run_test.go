package main

import (
	"sync"
	"testing"
	"time"
)

func TestRunHeartbeats(t *testing.T) {
	// This is the number of heartbeats that we are going to send.
	numberOfHeartbeats := 10

	// Create a channel that we'll use to publish heartbeat ticks.
	tickerChan := make(chan time.Time)
	mockTicker := time.Ticker{C: tickerChan}

	// Get a reference to the channel that stops the jobs
	var shutdownChannel chan error
	// We'll get a data race warning because "mockHeartBeatCheck" is invoked by another goroutine.
	var chanMutex sync.Mutex
	mockShutdownHandler := func(c chan error) {
		chanMutex.Lock()
		shutdownChannel = c
		chanMutex.Unlock()
	}

	// Count the number of checks we've performed. Each heartbeat should perform a check.
	checksPerformed := 0
	mockHeartBeatCheck := func() {
		checksPerformed++

		// When we've received the expected number of heartbeats we'll write
		// to the shutdownChannel which will unblock the run function.
		if checksPerformed == numberOfHeartbeats {
			chanMutex.Lock()
			shutdownChannel <- nil
			chanMutex.Unlock()
		}
	}

	ecg := ECG{
		check:     mockHeartBeatCheck,
		stopChan:  make(chan bool),
		heartbeat: &mockTicker,
	}

	// Publish heartbeat ticks.
	go func() {
		for i := 0; i < numberOfHeartbeats; i++ {
			tickerChan <- time.Now()
		}
	}()

	// This function blocks until something is published to the "shutdownChannel".
	err := run(mockShutdownHandler, &ecg)
	if err != nil {
		t.Errorf("Unexpected error :%s", err)
	}

	if checksPerformed != numberOfHeartbeats {
		t.Errorf("unexpected number of hearbeat checks; expected %d got %d", numberOfHeartbeats, checksPerformed)
	}
}
