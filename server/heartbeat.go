package server

import (
	"strconv"
	"time"
)

const (
	HeartbeatTTL      = time.Minute * 10
	heartbeatInterval = time.Second * 10
)

// CheckHeartbeat is used to check if the session has been inactive for more than
// ten minutes. If that is the case, the session will be terminated and saved to disk.
func (s *Server) checkHeartbeat() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.log.Debug("Checking heartbeat.",
		"last_heartbeat", s.lastHeartbeat,
		"time_now", s.clock.Now().UnixMilli(),
	)

	if s.activeBuffer == nil {
		return
	}

	if s.clock.Now().After(s.lastHeartbeat.Add(HeartbeatTTL)) {
		s.log.Info(
			"Writing the current buffer to disk due to inactivity.",
			"last_heartbeat", strconv.FormatInt(s.lastHeartbeat.UnixMilli(), 10),
			"current_time", strconv.FormatInt(s.clock.Now().UnixMilli(), 10),
			"end_time", strconv.FormatInt(s.lastHeartbeat.Add(HeartbeatTTL).UnixMilli(), 10),
		)
		s.saveBuffer()
	}
}

// runHeartbeatChecks runs in a separate goroutine and makes sure
// that no session is allowed to be idle for more than 10 minutes.
func (s *Server) runHeartbeatChecks() {
	go func() {
		ticker, stopTicker := s.clock.NewTicker(heartbeatInterval)
		defer stopTicker()
		for {
			select {
			case <-s.stopJobs:
				return
			case <-ticker:
				s.checkHeartbeat()
			}
		}
	}()
}
