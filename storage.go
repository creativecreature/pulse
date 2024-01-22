package codeharvest

type TemporaryStorage interface {
	Write(Session) error
	Read() (Sessions, error)
	Clean() error
}

type PermanentStorage interface {
	Write(s []AggregatedSession) error
	Aggregate(timeperiod TimePeriod) error
}
