package server

import (
	"context"
	"encoding/json"

	"github.com/creativecreature/pulse"
)

// writeToRemote will write the session to the remote storage.
func (s *Server) writeToRemote(session pulse.CodingSession) {
	if len(session.Repositories) == 0 {
		return
	}

	err := s.sessionWriter.Write(context.Background(), session)
	if err != nil {
		s.logger.Errorf("Failed to write the session to the permanent storage: %v", err)
	}
}

func (s *Server) aggregate() {
	s.mu.Lock()
	defer s.mu.Unlock()

	buffers := make(pulse.Buffers, 0)
	values := s.logDB.Aggregate()
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

func (s *Server) runAggregations(ctx context.Context) {
	go func() {
		ticker, stopTicker := s.clock.NewTicker(s.cfg.Server.AggregationInterval)
		defer stopTicker()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker:
				s.aggregate()
			}
		}
	}()
}
