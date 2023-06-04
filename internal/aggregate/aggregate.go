package aggregate

import (
	"time"

	"code-harvest.conner.dev/internal/storage"
	"code-harvest.conner.dev/internal/storage/models"
)

// day is used to group unix timestamps into days
func day(timestamp int64) int64 {
	var dayInMs int64 = 24 * 60 * 60 * 1000
	return timestamp - (timestamp % dayInMs)
}

// group groups the temporary sessions by day
func group(session []models.TemporarySession) map[int64][]models.TemporarySession {
	buckets := make(map[int64][]models.TemporarySession)
	for _, s := range session {
		d := day(s.StartedAt)
		buckets[d] = append(buckets[d], s)
	}
	return buckets
}

// repositores takes all the temporary sessions for a given day and returns all
// the repositories, and files within those repositories, that we've worked on.
func repositories(tempSessions []models.TemporarySession) []models.Repository {
	// The coding sessions may have opened files in different repositories.
	// Therefore, we'll have to aggregate all the files together by repository
	// and filename.
	repositoryFileMap := make(map[string]map[string]*models.AggregatedFile)
	for _, tempSession := range tempSessions {
		for _, file := range tempSession.Files {
			// Check if it is the first time we're seeing this repo. In that case
			// we'll initialize the map.
			if _, ok := repositoryFileMap[file.Repository]; !ok {
				repositoryFileMap[file.Repository] = make(map[string]*models.AggregatedFile)
			}

			// We got files for this repo. Let's check if it is the first time we're
			// seeing this file. In that case we can just go ahead and add it.
			if _, ok := repositoryFileMap[file.Repository][file.Name]; !ok {
				repositoryFileMap[file.Repository][file.Name] = &models.AggregatedFile{
					Name:       file.Name,
					Path:       file.Path,
					Filetype:   file.Filetype,
					DurationMs: file.DurationMs,
				}
				continue
			}

			// We've seen this file and repository before. We'll have to merge them.
			prevFile := repositoryFileMap[file.Repository][file.Name]
			prevFile.DurationMs += file.DurationMs
		}
	}

	// At this point we've merged all the files that could have been edited
	// during different sessions and saved them in our repository/filename map.
	// We can now turn the maps into a slice of repositories.
	repositories := make([]models.Repository, 0)
	for repositoryName, filePointers := range repositoryFileMap {
		var durationMs int64 = 0
		files := make([]models.AggregatedFile, 0)
		for _, filePointer := range filePointers {
			files = append(files, *filePointer)
			durationMs += filePointer.DurationMs
		}
		repository := models.Repository{
			Name:       repositoryName,
			Files:      files,
			DurationMs: durationMs,
		}
		repositories = append(repositories, repository)
	}
	return repositories
}

// sessions takes a map where the key is the day and the value is a slice of
// temporary sessions that have occurred during that day. It returns the
// aggregated sessions.
func sessions(buckets map[int64][]models.TemporarySession) []models.AggregatedSession {
	sessions := make([]models.AggregatedSession, 0)
	for date, tempSessions := range buckets {
		dateString := time.Unix(0, date*int64(time.Millisecond)).Format("2006-01-02")
		var totalTime int64 = 0
		for _, tempSession := range tempSessions {
			totalTime += tempSession.DurationMs
		}
		session := models.AggregatedSession{
			Date:         date,
			DateString:   dateString,
			TotalTimeMs:  totalTime,
			Repositories: repositories(tempSessions),
		}
		sessions = append(sessions, session)
	}
	return sessions
}

// AggregateSessions aggregates all of the sessions inside the tmp directory
func TemporarySessions() []models.AggregatedSession {
	tempStorage := storage.DiskStorage()
	tempSessions, err := tempStorage.GetAll()
	if err != nil {
		panic(err)
	}
	buckets := group(tempSessions)
	return sessions(buckets)
}
