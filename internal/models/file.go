package models

type File struct {
	OpenedAt   int64  `bson:"-"`
	ClosedAt   int64  `bson:"-"`
	Name       string `bson:"name"`
	Repository string `bson:"repository"`
	Path       string `bson:"path"`
	Filetype   string `bson:"filetype"`
	DurationMs int64  `bson:"duration_ms"`
}
