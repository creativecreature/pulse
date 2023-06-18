package aggregate

import "code-harvest.conner.dev/internal/storage"

// Day aggregates raw coding sessions by day and stores them in a database
func Day(tempStorage storage.TemporaryStorage, permStorage storage.PermanentStorage) error {
	sessions, err := tempStorage.GetAll()
	if err != nil {
		return err
	}
	err = permStorage.SaveAll(sessions.AggregateByDay())
	if err != nil {
		return err
	}
	return tempStorage.RemoveAll()
}
