package models

// NOTE: In this file, we have defined structs that correspond to the final
// format of our coding sessions. These structs are created by consolidating
// the coding sessions that we've temporarily stored to disk.

type AggregatedFile struct {
	Name       string
	Filetype   string
	DurationMs int64
}

// Repository represents all work that has been done in a repository during a day
type Repository struct {
	Name       string
	Files      []AggregatedFile
	DurationMs int64
}

// AggregatedSession is used to group all coding sessions that occurred during a day
type AggregatedSession struct {
	Date         int64
	DateString   string // yyyy-mm-dd
	TotalTimeMs  int64
	Repositories []Repository
}
