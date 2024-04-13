package pulse

// TemporaryStorage is an abstraction for a storage that can be used to store
// the session data temporarily. Every editor instance is going to have its own
// session. Therefore, it's common to have several sessions for any given day.
// To not exceed any database free tier limits, the sessions are first stored
// temporarily, and can then aggregated to a permanent storage.
type TemporaryStorage interface {
	Write(Session) error
	Read() (Sessions, error)
	Clean() error
}

// PermanentStorage is an abstraction for a storage that allows
// you to aggregate sessions from the temporary storage.
type PermanentStorage interface {
	Write(s []AggregatedSession) error
	Aggregate(timeperiod Period) error
}
