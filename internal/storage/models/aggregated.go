package models

// NOTE: In this file, we have defined structs that correspond to the final
// format of our coding sessions. These structs are created by consolidating
// the coding sessions that we've temporarily stored to disk.

type AggregatedFile struct {
	Name       string `bson:"name"`
	Path       string `bson:"path"`
	Filetype   string `bson:"filetype"`
	DurationMs int64  `bson:"duration_ms"`
}

// Repository represents all work that has been done in a repository during a day
type Repository struct {
	Name       string           `bson:"name"`
	Files      []AggregatedFile `bson:"files"`
	DurationMs int64            `bson:"duration_ms"`
}

// AggregatedSession is used to group all coding sessions that occurred during a day
type AggregatedSession struct {
	ID           string       `bson:"_id,omitempty"`
	Date         int64        `bson:"date"`
	DateString   string       `bson:"date_string"` // yyyy-mm-dd
	TotalTimeMs  int64        `bson:"total_time_ms"`
	Repositories []Repository `bson:"repositories"`
}
