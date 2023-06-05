package domain

// Repository represents all work that has been done in a repository during a day
type Repository struct {
	Name       string      `bson:"name"`
	Files      []DailyFile `bson:"files"`
	DurationMs int64       `bson:"duration_ms"`
}
