package mongo

import "context"

func (m *db) deleteByDateRange(minDate, maxDate int64) error {
	filter := createDateFilter(minDate, maxDate)
	_, err := m.client.Database(m.database).
		Collection(daily).
		DeleteMany(context.Background(), filter)
	return err
}

func (m *db) deleteCollection(collection string) error {
	return m.client.Database(m.database).Collection(collection).Drop(context.Background())
}
