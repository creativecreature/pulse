package domain

type AggregatedSessions []AggregatedSession

func (previousSessions AggregatedSessions) Merge(newSessions AggregatedSessions) AggregatedSessions {
	datePrevSession := make(map[string]AggregatedSession)
	for _, prevSession := range previousSessions {
		datePrevSession[prevSession.DateString] = prevSession
	}
	mergedSessions := make([]AggregatedSession, 0)
	for _, newSession := range newSessions {
		// Check if we should merge this with a previous session
		if prevSession, ok := datePrevSession[newSession.DateString]; ok {
			repositories := prevSession.Repositories.Merge(newSession.Repositories)
			session := AggregatedSession{
				ID:           prevSession.ID,
				Period:       prevSession.Period,
				Date:         newSession.Date,
				DateString:   newSession.DateString,
				TotalTimeMs:  prevSession.TotalTimeMs + newSession.TotalTimeMs,
				Repositories: repositories,
			}
			mergedSessions = append(mergedSessions, session)
			continue
		}
		// If this is the first session for the given date we'll just append it to
		// the slice
		mergedSessions = append(mergedSessions, newSession)
	}
	return mergedSessions
}
