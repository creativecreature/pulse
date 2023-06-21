package mongo

import (
	"context"
	"math"
	"time"

	"code-harvest.conner.dev/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func createDateSortOptions() *options.FindOptions {
	return options.Find().SetSort(bson.D{{Key: "date", Value: 1}})
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
	minDate = math.MaxInt64
	maxDate = math.MinInt64
	for _, s := range sessions {
		minDate, maxDate = min(minDate, s.Date), max(maxDate, s.Date)
	}
	return minDate, maxDate
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

func (m *db) getByDateRange(minDate, maxDate int64) (domain.AggregatedSessions, error) {
	filter := createDateFilter(minDate, maxDate)
	sortOptions := createDateSortOptions()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := m.client.Database(m.database).Collection(daily).Find(ctx, filter, sortOptions)
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

func (m *db) readAll() (domain.AggregatedSessions, error) {
	sortOptions := createDateSortOptions()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := m.client.Database(m.database).Collection(daily).Find(ctx, bson.M{}, sortOptions)
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
