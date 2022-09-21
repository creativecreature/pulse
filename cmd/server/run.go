package main

type StartStop interface {
	start() error
	stop()
}

// Blocks until we receive an error from any of the
// startStoppers or a shutdown signal from the OS.
func run(shutdownFunc func(chan error), startStoppers ...StartStop) error {
	shutdownChan := make(chan error)

	go shutdownFunc(shutdownChan)

	for _, s := range startStoppers {
		startStopper := s
		go func() {
			shutdownChan <- startStopper.start()
		}()
	}

	// Blocks until we receive anything on the channel
	shutdownErr := <-shutdownChan
	for _, s := range startStoppers {
		s.stop()
	}

	return shutdownErr
}
