package domain

type AggregatedFile struct {
	Name       string `bson:"name"`
	Path       string `bson:"path"`
	Filetype   string `bson:"filetype"`
	DurationMs int64  `bson:"duration_ms"`
}

func (a AggregatedFile) merge(b AggregatedFile) AggregatedFile {
	return AggregatedFile{
		Name:       a.Name,
		Path:       a.Path,
		Filetype:   a.Filetype,
		DurationMs: a.DurationMs + b.DurationMs,
	}
}
