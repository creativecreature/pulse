package mongo

import (
	"context"

	"code-harvest.conner.dev/domain"
)

func (m *db) insertAll(collection string, sessions []domain.AggregatedSession) error {
	documents := make([]interface{}, 0)
	for _, session := range sessions {
		documents = append(documents, session)
	}
	_, err := m.client.Database(m.database).
		Collection(collection).
		InsertMany(context.Background(), documents)
	return err
}

// Write writes daily coding sessions to a mongodb collection
func (m *db) Write(sessions []domain.AggregatedSession) error {
	minDate, maxDate := dateRange(sessions)
	previousSessionsForRange, err := m.getByDateRange(minDate, maxDate)
	if err != nil {
		return err
	}
	// There were no previous sessions for this range of dates
	if len(previousSessionsForRange) == 0 {
		return m.insertAll(daily, sessions)
	}

	// We have already aggregated sessions for this day. We'll have to merge them.
	combinedSessions := make(domain.AggregatedSessions, 0)
	combinedSessions = append(combinedSessions, previousSessionsForRange...)
	combinedSessions = append(combinedSessions, sessions...)
	mergedSessions := combinedSessions.MergeByDay()

	// Delete the previously stored sessions for this range
	err = m.deleteByDateRange(minDate, maxDate)
	if err != nil {
		return err
	}

	// Update this range of sessions with the merged ones
	return m.insertAll(daily, mergedSessions)
}
