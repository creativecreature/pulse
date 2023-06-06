package domain

// AggregatedFile represents all the work that has been done in a patricular file for a
// given day
type AggregatedFile struct {
	Name       string `bson:"name"`
	Path       string `bson:"path"`
	Filetype   string `bson:"filetype"`
	DurationMs int64  `bson:"duration_ms"`
}
