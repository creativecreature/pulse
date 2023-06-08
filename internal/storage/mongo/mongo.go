package mongo

import (
	"context"
	"time"

	"code-harvest.conner.dev/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type db struct {
	uri        string
	database   string
	collection string
	client     *mongo.Client
}

func NewDB(uri, database, collection string) *db {
	return &db{
		uri:        uri,
		database:   database,
		collection: collection,
	}
}

func (m *db) Connect() func() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(m.uri))
	// Can't proceed without a database connection.
	if err != nil {
		panic(err)
	}

	m.client = client

	return func() {
		err := client.Disconnect(ctx)
		if err != nil {
			panic(err)
		}
	}
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func dateRange(sessions []domain.AggregatedSession) (minDate, maxDate int64) {
	for _, s := range sessions {
		minDate, maxDate = min(minDate, s.Date), max(maxDate, s.Date)
	}
	return minDate, maxDate
}

func (m *db) getByDateRange(minDate, maxDate int64) (domain.AggregatedSessions, error) {
	filter := bson.D{
		{
			Key: "$and",
			Value: bson.A{
				bson.D{{Key: "date", Value: bson.D{{Key: "$gte", Value: minDate}}}},
				bson.D{{Key: "date", Value: bson.D{{Key: "$lte", Value: maxDate}}}},
			},
		},
	}
	sort := bson.D{{Key: "date", Value: 1}}
	opts := options.Find().SetSort(sort)
	cursor, err := m.client.Database(m.database).
		Collection(m.collection).
		Find(context.Background(), filter, opts)
	if err != nil {
		return domain.AggregatedSessions{}, err
	}

	results := make([]domain.AggregatedSession, 0)
	err = cursor.All(context.Background(), &results)
	if err != nil {
		return domain.AggregatedSessions{}, err
	}
	return results, nil
}

func (m *db) deleteByDateRange(minDate, maxDate int64) error {
	filter := bson.D{
		{
			Key: "$and",
			Value: bson.A{
				bson.D{{Key: "date", Value: bson.D{{Key: "$gte", Value: minDate}}}},
				bson.D{{Key: "date", Value: bson.D{{Key: "$lte", Value: maxDate}}}},
			},
		},
	}
	_, err := m.client.Database(m.database).
		Collection(m.collection).
		DeleteMany(context.Background(), filter)
	return err
}

func (m *db) insertAll(sessions []domain.AggregatedSession) error {
	documents := make([]interface{}, 0)
	for _, session := range sessions {
		documents = append(documents, session)
	}
	_, err := m.client.Database(m.database).
		Collection(m.collection).
		InsertMany(context.Background(), documents)
	return err
}

func (m *db) SaveAll(sessions []domain.AggregatedSession) error {
	minDate, maxDate := dateRange(sessions)
	previousSessionsForRange, err := m.getByDateRange(minDate, maxDate)
	if err != nil {
		return err
	}
	// There were no previous sessions for this range of dates
	if len(previousSessionsForRange) == 0 {
		return m.insertAll(sessions)
	}

	// Merge the new sessions with the previous ones
	mergedSessions := previousSessionsForRange.Merge(sessions)

	// Delete the previously stored sessions for this range
	err = m.deleteByDateRange(minDate, maxDate)
	if err != nil {
		return err
	}

	// Update this range of sessions with the merged ones
	return m.insertAll(mergedSessions)
}
