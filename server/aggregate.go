package server

import (
	"context"
	"encoding/json"
	"time"

	"github.com/creativecreature/pulse"
)

const aggregationInterval = 15 * time.Minute

func (s *Server) createSession(buffers pulse.Buffers) pulse.AggregatedSession {
	repos := make(map[string]pulse.Repository)
	for _, buf := range buffers {
		repo, ok := repos[buf.Repository]
		if !ok {
			repos[buf.Repository] = pulse.Repository{
				Name:  buf.Repository,
				Files: make(pulse.AggregatedFiles, 0),
			}
		}

		file := pulse.AggregatedFile{
			Name:       buf.Filename,
			Path:       buf.Filepath,
			Filetype:   buf.Filetype,
			DurationMs: buf.Duration.Milliseconds(),
		}
		repo.DurationMs += file.DurationMs
		repo.Files = append(repo.Files, file)
		repos[buf.Repository] = repo
	}

	var totalDurationMS int64
	repositories := make(pulse.Repositories, 0, len(repos))
	for _, repo := range repos {
		totalDurationMS += repo.DurationMs
		repositories = append(repositories, repo)
	}

	session := pulse.AggregatedSession{
		Period:       pulse.Day,
		EpochDateMs:  s.clock.Now().Truncate(time.Millisecond).UnixMilli(),
		DateString:   s.clock.Now().Format("2006-01-02"),
		TotalTimeMs:  totalDurationMS,
		Repositories: repositories,
	}
	return session
}

// syncWithRemote will sync the local database with the remote database.
func (s *Server) syncWithRemote(session pulse.AggregatedSession) {
	if len(session.Repositories) == 0 {
		return
	}

	err := s.permanentStorage.Write(context.Background(), session)
	if err != nil {
		s.log.Errorf("Failed to write the session to the permanent storage: %v", err)
	}
}

func (s *Server) aggregate() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	buffers := make(pulse.Buffers, 0)
	values := s.localStorage.Aggregate()
	for _, value := range values {
		var buf pulse.Buffer
		err := json.Unmarshal(value, &buf)
		if err != nil {
			panic(err)
		}
		buffers = append(buffers, buf)
	}
	aggregatedSession := s.createSession(buffers)
	go s.syncWithRemote(aggregatedSession)
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
