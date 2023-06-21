package mongo

import (
	"errors"

	"code-harvest.conner.dev/domain"
)

func (m *db) Aggregate(timePeriod domain.TimePeriod) error {
	dailySessions, err := m.readAll()
	if err != nil {
		return err
	}

	sessions, collection := domain.AggregatedSessions{}, ""
	switch tPeriod := timePeriod; tPeriod {
	case domain.Week:
		sessions = dailySessions.MergeByWeek()
		collection = weekly
	case domain.Month:
		sessions = dailySessions.MergeByMonth()
		collection = monthly
	case domain.Year:
		sessions = dailySessions.MergeByYear()
		collection = yearly
	}

	if len(sessions) == 0 {
		return errors.New("no sessions to aggregate for the given time period")
	}

	err = m.deleteCollection(collection)
	if err != nil {
		return err
	}

	return m.insertAll(collection, sessions)
}
