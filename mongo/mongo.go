package mongo

import (
	"context"
	"math"
	"time"

	"github.com/charmbracelet/log"
	"github.com/creativecreature/pulse"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Constants for the collection names.
const (
	collectionDaily   = "daily"
	collectionWeekly  = "weekly"
	collectionMonthly = "monthly"
	collectionYearly  = "yearly"
)

type Client struct {
	*mongo.Client
	database string
	log      *log.Logger
}

func New(uri, database string, log *log.Logger) *Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	return &Client{
		Client:   client,
		database: database,
		log:      log,
	}
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

func dateRange(sessions []pulse.AggregatedSession) (minDate, maxDate int64) {
	minDate, maxDate = math.MaxInt64, math.MinInt64
	for _, s := range sessions {
		minDate = min(minDate, s.EpochDateMs)
		maxDate = max(maxDate, s.EpochDateMs)
	}
	return minDate, maxDate
}

func (c *Client) getByDateRange(ctx context.Context, minDate, maxDate int64) (pulse.AggregatedSessions, error) {
	filter := createDateFilter(minDate, maxDate)
	dateSortOpts := options.Find().SetSort(bson.D{{Key: "date", Value: 1}})
	cursor, err := c.Database(c.database).Collection(collectionDaily).Find(ctx, filter, dateSortOpts)
	if err != nil {
		return pulse.AggregatedSessions{}, err
	}

	results := make([]pulse.AggregatedSession, 0)
	err = cursor.All(ctx, &results)
	if err != nil {
		return pulse.AggregatedSessions{}, err
	}

	return results, nil
}

func (c *Client) deleteByDateRange(ctx context.Context, minDate, maxDate int64) error {
	filter := createDateFilter(minDate, maxDate)
	_, err := c.Database(c.database).
		Collection(collectionDaily).
		DeleteMany(ctx, filter)
	return err
}

func (c *Client) insertAll(ctx context.Context, collection string, sessions []pulse.AggregatedSession) error {
	documents := make([]interface{}, 0)
	for _, session := range sessions {
		documents = append(documents, session)
	}

	_, err := c.Database(c.database).Collection(collection).InsertMany(ctx, documents)
	return err
}

func (c *Client) readAll(ctx context.Context) (pulse.AggregatedSessions, error) {
	sortOpts := options.Find().SetSort(bson.D{{Key: "date", Value: 1}})
	cursor, err := c.Database(c.database).Collection(collectionDaily).Find(ctx, bson.M{}, sortOpts)
	if err != nil {
		return pulse.AggregatedSessions{}, err
	}

	results := make([]pulse.AggregatedSession, 0)
	err = cursor.All(ctx, &results)
	if err != nil {
		return pulse.AggregatedSessions{}, err
	}

	return results, nil
}

func (c *Client) aggregate(ctx context.Context) error {
	dailySessions, err := c.readAll(ctx)
	if err != nil {
		return err
	}

	// Aggregate by week.
	c.log.Info("Dropping the previous aggregation for this week.")
	err = c.Database(c.database).Collection(collectionWeekly).Drop(ctx)
	if err != nil {
		return err
	}
	c.log.Info("Generating a new weekly aggregation.")
	err = c.insertAll(ctx, collectionWeekly, dailySessions.MergeByWeek())
	if err != nil {
		return err
	}

	// Aggregate by month.
	c.log.Info("Dropping the previous aggregation for this month.")
	err = c.Database(c.database).Collection(collectionMonthly).Drop(ctx)
	if err != nil {
		return err
	}
	c.log.Info("Generating a new monthly aggregation.")
	err = c.insertAll(ctx, collectionMonthly, dailySessions.MergeByMonth())
	if err != nil {
		return err
	}

	// Aggregate by year.
	c.log.Info("Dropping the previous aggregation for this year.")
	err = c.Database(c.database).Collection(collectionYearly).Drop(ctx)
	if err != nil {
		return err
	}
	c.log.Info("Generating a new yearly aggregation.")
	return c.insertAll(ctx, collectionYearly, dailySessions.MergeByYear())
}

// Write writes daily coding sessions to a mongodb collection.
func (c *Client) Write(ctx context.Context, sessions []pulse.AggregatedSession) error {
	// We might aggregate sessions from the temp storage several times a
	// day. Therefore, we have to fetch any previous sessions for the same
	// timeframe. If we have any, we'll merge them with the new ones.
	minDate, maxDate := dateRange(sessions)
	previousSessionsForRange, err := c.getByDateRange(ctx, minDate, maxDate)
	if err != nil {
		return err
	}

	// If there were no previous sessions for this range of dates, we'll simply insert them.
	if len(previousSessionsForRange) == 0 {
		c.log.Info("Inserting as is because no previous session have been aggregated for this day.")
		return c.insertAll(ctx, collectionDaily, sessions)
	}

	// If we reach this point, it means that we've aggregated sessions for this
	// day before. We now have to go through the process of merging them.
	c.log.Info("Merging the disk sessions with the previously aggregated session for this day.")
	combinedSessions := make(pulse.AggregatedSessions, 0, len(previousSessionsForRange)+len(sessions))
	combinedSessions = append(combinedSessions, previousSessionsForRange...)
	combinedSessions = append(combinedSessions, sessions...)
	mergedSessions := combinedSessions.MergeByDay()

	// Delete the previously stored sessions for this range
	c.log.Info("Deleting the previously aggregated session for this day.")
	err = c.deleteByDateRange(ctx, minDate, maxDate)
	if err != nil {
		return err
	}

	// Update this range of sessions with the result of the merger.
	c.log.Info("Inserting the result of the merger.")
	err = c.insertAll(ctx, collectionDaily, mergedSessions)
	if err != nil {
		return err
	}

	// Lastly, we'll update the aggregated collections to be
	// able to display the data per week, month, and year.
	return c.aggregate(ctx)
}
