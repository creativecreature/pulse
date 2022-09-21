package main

import (
	"errors"
	"time"
)

// ECG is responsible for running a heartbeat check.
// The check will look at the session and look at the
// timestamp for the last action. If its been more than
// x minutes I will end the session.
type ECG struct {
	check     func()
	stopChan  chan bool
	heartbeat *time.Ticker
}

func (ecg *ECG) start() error {
	if ecg.check == nil {
		return errors.New("the ECG requires a check function")
	}

	for {
		select {
		case <-ecg.heartbeat.C:
			ecg.check()
		case <-ecg.stopChan:
			return nil
		}
	}
}

func (ecg *ECG) stop() {
	ecg.heartbeat.Stop()
}
