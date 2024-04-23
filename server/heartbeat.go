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
	s.log.Debug("Checking heartbeat.",
		"active_editor_id", s.activeEditorID,
		"last_heartbeat", s.lastHeartbeat,
		"time_now", s.clock.GetTime(),
	)
	if s.activeEditorID == "" {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.lastHeartbeat+HeartbeatTTL.Milliseconds() < s.clock.GetTime() {
		s.log.Info(
			"Ending all active sessions due to inactivity",
			"last_heartbeat", strconv.FormatInt(s.lastHeartbeat, 10),
			"current_time", strconv.FormatInt(s.clock.GetTime(), 10),
			"end_time", strconv.FormatInt(s.lastHeartbeat+int64(HeartbeatTTL), 10),
		)

		// The machine may have entered sleep mode, potentially stopping the heartbeat
		// check from executing at its scheduled interval. To mitigate this, the session
		// will be terminated based on the time of the last recorded heartbeat plus the
		// TTL. This prevents the creation of inaccurately long sessions.
		s.saveAllSessions(s.lastHeartbeat + int64(HeartbeatTTL/time.Millisecond))
		s.activeEditorID = ""
	}
}

// startHeartbeatChecks runs in a separate goroutine and makes sure
// that no session is allowed to be idle for more than 10 minutes.
func (s *Server) startHeartbeatChecks() {
	go func() {
		ticker, stopTicker := s.clock.NewTicker(heartbeatInterval)
		defer stopTicker()
		for {
			select {
			case <-s.stopHeartbeatChecks:
				return
			case <-ticker:
				s.checkHeartbeat()
			}
		}
	}()
}
