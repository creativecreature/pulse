package mongo

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/creativecreature/code-harvest"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Constants for the collection names
const (
	daily   = "daily"
	weekly  = "weekly"
	monthly = "monthly"
	yearly  = "yearly"
)

type db struct {
	uri      string
	database string
	client   *mongo.Client
}

func NewDB(uri, database string) *db {
	return &db{
		uri:      uri,
		database: database,
	}
}

func (m *db) Connect() func() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(m.uri))
	// We should be able to connect to the database. If we can't there isn't much
	// that we are able to do.
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

func createDateFilter(minDate, maxDate int64) primitive.D {
	return bson.D{
		{
			Key: "$and",
			Value: bson.A{
				bson.D{{Key: "date", Value: bson.D{{Key: "$gte", Value: minDate}}}},
				bson.D{{Key: "date", Value: bson.D{{Key: "$lte", Value: maxDate}}}},
			},
		},
	}
}

func createDateSortOptions() *options.FindOptions {
	return options.Find().SetSort(bson.D{{Key: "date", Value: 1}})
}

func dateRange(sessions []codeharvest.AggregatedSession) (minDate, maxDate int64) {
	minDate = math.MaxInt64
	maxDate = math.MinInt64
	for _, s := range sessions {
		minDate, maxDate = min(minDate, s.Date), max(maxDate, s.Date)
	}
	return minDate, maxDate
}

func (m *db) getByDateRange(minDate, maxDate int64) (codeharvest.AggregatedSessions, error) {
	filter := createDateFilter(minDate, maxDate)
	sortOptions := createDateSortOptions()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := m.client.Database(m.database).
		Collection(daily).
		Find(ctx, filter, sortOptions)
	if err != nil {
		return codeharvest.AggregatedSessions{}, err
	}

	results := make([]codeharvest.AggregatedSession, 0)
	err = cursor.All(context.Background(), &results)
	if err != nil {
		return codeharvest.AggregatedSessions{}, err
	}

	return results, nil
}

func (m *db) deleteByDateRange(minDate, maxDate int64) error {
	filter := createDateFilter(minDate, maxDate)
	_, err := m.client.Database(m.database).
		Collection(daily).
		DeleteMany(context.Background(), filter)
	return err
}

func (m *db) insertAll(collection string, sessions []codeharvest.AggregatedSession) error {
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
func (m *db) Write(sessions []codeharvest.AggregatedSession) error {
	// We might aggregate sessions from the temp storage several times a day.
	// Therefore, we have to fetch any previous sessions for the same timeframe
	// and if we have any we'll have to merge them with the new ones.
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
	combinedSessions := make(codeharvest.AggregatedSessions, 0)
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

func (m *db) readAll() (codeharvest.AggregatedSessions, error) {
	sortOptions := createDateSortOptions()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := m.client.Database(m.database).
		Collection(daily).
		Find(ctx, bson.M{}, sortOptions)
	if err != nil {
		return codeharvest.AggregatedSessions{}, err
	}

	results := make([]codeharvest.AggregatedSession, 0)
	err = cursor.All(context.Background(), &results)
	if err != nil {
		return codeharvest.AggregatedSessions{}, err
	}

	return results, nil
}

func (m *db) deleteCollection(collection string) error {
	return m.client.Database(m.database).
		Collection(collection).
		Drop(context.Background())
}

func (m *db) Aggregate(timePeriod codeharvest.TimePeriod) error {
	if timePeriod == codeharvest.Day {
		return errors.New("cannot aggregate by day")
	}

	dailySessions, err := m.readAll()
	if err != nil {
		return err
	}

	sessions, collection := codeharvest.AggregatedSessions{}, ""
	switch tPeriod := timePeriod; tPeriod {
	case codeharvest.Week:
		sessions = dailySessions.MergeByWeek()
		collection = weekly
	case codeharvest.Month:
		sessions = dailySessions.MergeByMonth()
		collection = monthly
	case codeharvest.Year:
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
