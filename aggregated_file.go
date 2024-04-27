package pulse

import "cmp"

// AggregatedFile represents a file that has been aggregated
// for a given time period (day, week, month, year).
type AggregatedFile struct {
	Name       string `bson:"name"`
	Path       string `bson:"path"`
	Filetype   string `bson:"filetype"`
	DurationMs int64  `bson:"duration_ms"`
}

// merge takes two AggregatedFile, merges them, and returns the result.
func (a AggregatedFile) merge(b AggregatedFile) AggregatedFile {
	return AggregatedFile{
		Name:       cmp.Or(a.Name, b.Name),
		Path:       cmp.Or(a.Path, b.Path),
		Filetype:   cmp.Or(a.Filetype, b.Filetype),
		DurationMs: a.DurationMs + b.DurationMs,
	}
}
