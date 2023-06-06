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

func (m *db) getByDateRange(minDate, maxDate int64) ([]domain.AggregatedSession, error) {
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
		return []domain.AggregatedSession{}, err
	}

	results := make([]domain.AggregatedSession, 0)
	err = cursor.All(context.Background(), &results)
	if err != nil {
		return []domain.AggregatedSession{}, err
	}
	return results, nil
}

func mergeFiles(prevFiles, newFiles []domain.AggregatedFile) []domain.AggregatedFile {
	prevFilesMap := make(map[string]domain.AggregatedFile)
	newFilesMap := make(map[string]domain.AggregatedFile)
	for _, file := range prevFiles {
		prevFilesMap[file.Path] = file
	}
	for _, file := range newFiles {
		newFilesMap[file.Path] = file
	}

	mergedFiles := make([]domain.AggregatedFile, 0)
	for _, prevFile := range prevFiles {
		// This file haven't been worked on in the new session. We'll just
		// add it to the final slice
		newFile, ok := newFilesMap[prevFile.Path]
		if !ok {
			mergedFiles = append(mergedFiles, prevFile)
			continue
		}

		mergedFile := domain.AggregatedFile{
			Name:       prevFile.Name,
			Path:       prevFile.Path,
			Filetype:   prevFile.Filetype,
			DurationMs: prevFile.DurationMs + newFile.DurationMs,
		}
		mergedFiles = append(mergedFiles, mergedFile)
	}

	for _, newFile := range newFiles {
		// We have already handled the merging in the loop above. Here we'll just
		// add the new file which haven't been worked on in the previous session.
		if _, ok := prevFilesMap[newFile.Path]; !ok {
			mergedFiles = append(mergedFiles, newFile)
		}
	}

	return mergedFiles
}

func mergeRepositories(previousRepositories, newRepositories []domain.Repository) []domain.Repository {
	prevReposMap := make(map[string]domain.Repository)
	newReposMap := make(map[string]domain.Repository)
	for _, repository := range previousRepositories {
		prevReposMap[repository.Name] = repository
	}
	for _, repository := range newRepositories {
		newReposMap[repository.Name] = repository
	}

	mergedRepositories := make([]domain.Repository, 0)
	for _, prevRepo := range previousRepositories {
		// This repository haven't been worked on in the new session. We'll just
		// add it to the final slice
		newRepo, ok := newReposMap[prevRepo.Name]
		if !ok {
			mergedRepositories = append(mergedRepositories, prevRepo)
			continue
		}

		// This repository has been worked on in both sessions. We'll have to merge them
		mergedFiles := mergeFiles(prevRepo.Files, newRepo.Files)
		mergedRepository := domain.Repository{
			Name:       prevRepo.Name,
			DurationMs: prevRepo.DurationMs + newRepo.DurationMs,
			Files:      mergedFiles,
		}
		mergedRepositories = append(mergedRepositories, mergedRepository)
	}

	for _, newRepo := range newRepositories {
		// We have already handled the merging in the loop above. Here we'll just
		// add the new repositories which haven't been worked on in the previous
		// session.
		if _, ok := prevReposMap[newRepo.Name]; !ok {
			mergedRepositories = append(mergedRepositories, newRepo)
		}
	}

	return mergedRepositories
}

// mergeWithPreviousSessions merges the new sessions with old sessions that
// have occurred during the same day
func mergeWithPreviousSessions(previousSessions, newSessions []domain.AggregatedSession) []domain.AggregatedSession {
	datePrevSession := make(map[string]domain.AggregatedSession)
	for _, prevSession := range previousSessions {
		datePrevSession[prevSession.DateString] = prevSession
	}
	mergedSessions := make([]domain.AggregatedSession, 0)
	for _, newSession := range newSessions {
		// Check if we should merge this with a previous session
		if prevSession, ok := datePrevSession[newSession.DateString]; ok {
			repositories := mergeRepositories(prevSession.Repositories, newSession.Repositories)
			session := domain.AggregatedSession{
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
	mergedSessions := mergeWithPreviousSessions(previousSessionsForRange, sessions)

	// Delete the previously stored sessions for this range
	err = m.deleteByDateRange(minDate, maxDate)
	if err != nil {
		return err
	}

	// Update this range of sessions with the merged ones
	return m.insertAll(mergedSessions)
}
