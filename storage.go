package pulse

import "context"

// PermanentStorage is an abstraction for a storage that allows you to store sessions permanently.
type PermanentStorage interface {
	Write(context.Context, AggregatedSession) error
}
