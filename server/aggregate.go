package server

import (
	"context"
	"encoding/json"
	"time"

	"github.com/creativecreature/pulse"
)

const aggregationInterval = 15 * time.Minute

// writeToRemote will write the session to the remote storage.
func (s *Server) writeToRemote(session pulse.CodingSession) {
	if len(session.Repositories) == 0 {
		return
	}

	err := s.sessionWriter.Write(context.Background(), session)
	if err != nil {
		s.log.Errorf("Failed to write the session to the permanent storage: %v", err)
	}
}

func (s *Server) aggregate() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	buffers := make(pulse.Buffers, 0)
	values := s.db.Aggregate()
	for _, value := range values {
		var buf pulse.Buffer
		err := json.Unmarshal(value, &buf)
		if err != nil {
			panic(err)
		}
		buffers = append(buffers, buf)
	}
	codingSession := pulse.NewCodingSession(buffers, s.clock.Now())
	go s.writeToRemote(codingSession)
}

func (s *Server) runAggregations() {
	go func() {
		ticker, stopTicker := s.clock.NewTicker(aggregationInterval)
		defer stopTicker()
		for {
			select {
			case <-s.stopJobs:
				return
			case <-ticker:
				s.aggregate()
			}
		}
	}()
}
