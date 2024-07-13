package logdb

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/creativecreature/pulse/clock"
	"github.com/creativecreature/pulse/logger"
)

// Record represents a key-value pair in our database.
type Record struct {
	Key   string `json:"key"`
	Value []byte `json:"value"`
}

// LogDB is a simple key-value store that persists data to a log file.
type LogDB struct {
	sync.RWMutex
	dirPath          string
	segmentSizeBytes int64
	clock            clock.Clock
	log              *log.Logger
	head             *Segment
	tail             *Segment
}

// NewDB creates a new log database.
func NewDB(dirPath string, segmentSizeKB int, c clock.Clock) *LogDB {
	log := logger.New()

	// Create the directory if it doesn't exist.
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		log.Fatal(err)
	}

	segmentPaths, err := getSegmentPaths(dirPath)
	if err != nil {
		log.Fatal("could not get segment paths")
	}

	var logDB LogDB
	logDB.dirPath = dirPath
	logDB.segmentSizeBytes = int64(segmentSizeKB) * 1024
	logDB.log = log
	logDB.clock = c

	// If the directory is empty, we'll simply create the initial segment and return.
	if len(segmentPaths) == 0 {
		segment := newSegment(dirPath, 0)
		logDB.head, logDB.tail = segment, nil
		return &logDB
	}

	// Restore the previous segments.
	segments := restoreSegments(segmentPaths)

	var tail *Segment
	if len(segments) > 1 {
		tail = segments[len(segments)-1]
	}
	logDB.head, logDB.tail = segments[0], tail

	return &logDB
}

// RunSegmentations starts the database's compaction process.
func (db *LogDB) RunSegmentations(ctx context.Context, segmentationInterval time.Duration) {
	c, cancel := db.clock.NewTicker(segmentationInterval)
	defer cancel()
	for {
		select {
		case <-c:
			db.compact()
		case <-ctx.Done():
			return
		}
	}
}

// appendSegment creates a new segment and appends it to the
// head of the linked list. should be called with a lock.
func (db *LogDB) appendSegment() {
	db.log.Info("Appending a new segment")

	nextSegmentIndex := db.head.index + 1
	segment := newSegment(db.dirPath, nextSegmentIndex)

	if db.tail == nil {
		segment.next, segment.prev = db.head, db.head
		db.head.prev, db.head.next = segment, segment
		db.head, db.tail = segment, db.head
		return
	}

	segment.next, segment.prev = db.head, db.tail
	db.head.prev, db.tail.next = segment, segment
	db.head = segment
}

// compact compacts all of the segments together, removing any duplicate keys.
func (db *LogDB) compact() {
	db.Lock()
	defer db.Unlock()

	currentHead, currentTail := db.head, db.tail
	current := currentTail
	if current == nil {
		db.log.Info("Not enough segments to necessitate a compaction")
		return
	}

	db.log.Info("Compacting segments")
	valuesToWrite := make(map[string][]byte)

	for current != nil {
		current.Lock()
		for key := range current.hashIndex {
			var found bool
			for cursor := currentHead; cursor != current; cursor = cursor.next {
				_, found = cursor.getNoLock(key)
				if found {
					break
				}
			}

			// If this key was unique for all previous segments, we'll write it to the head.
			if !found {
				bytes, _ := current.getNoLock(key)
				valuesToWrite[key] = bytes
			}
		}

		// Adjust pointers to remove the current segment
		if current == currentHead {
			// Only one segment left, no need to delete it
			current.Unlock()
			break
		}

		prev := current.prev
		next := current.next

		if prev != nil {
			prev.next = next
		}
		if next != nil {
			next.prev = prev
		}

		if current == db.tail {
			db.tail = prev
		}
		if current == db.head {
			db.head = next
		}

		if err := current.delete(); err != nil {
			db.log.Error(err)
			current.Unlock()
			return
		}

		nextSegment := prev
		current.Unlock()
		current = nextSegment
	}

	if db.head == db.tail {
		db.tail = nil
	}

	for key, value := range valuesToWrite {
		db.mustSet(key, value)
	}
	db.log.Info("Finished compacting segments")
}

// Get retrieves a value from the database.
func (db *LogDB) Get(key string) ([]byte, bool) {
	db.RLock()
	defer db.RUnlock()

	current, head := db.head, db.head
	for {
		if value, ok := current.get(key); ok {
			return value, true
		}

		current = current.next

		if current == nil || current == head {
			break
		}
	}
	return nil, false
}

func (db *LogDB) GetAllUnique() map[string][]byte {
	db.Lock()
	defer db.Unlock()

	values := make(map[string][]byte, len(db.head.hashIndex))
	current := db.head
	for {
		for key := range current.hashIndex {
			if _, ok := values[key]; !ok {
				value, _ := current.get(key)
				values[key] = value
			}
		}

		// Update current and break if we've reached the tail.
		if current.next == db.head || current.next == nil {
			break
		}

		current = current.next
	}
	return values
}

// Set writes a key-value pair to the log file.
func (db *LogDB) Set(key string, value []byte) error {
	db.Lock()
	defer db.Unlock()

	err := db.head.set(key, value)
	if err != nil {
		return err
	}
	if db.head.size() >= db.segmentSizeBytes {
		db.appendSegment()
	}
	return nil
}

// Set writes a key-value pair without locking the database.
func (db *LogDB) set(key string, value []byte) error {
	err := db.head.set(key, value)
	if err != nil {
		return err
	}
	if db.head.size() >= db.segmentSizeBytes {
		db.appendSegment()
	}
	return nil
}

// MustSet writes a key-value pair to the log file and panics on error.
func (db *LogDB) MustSet(key string, value []byte) {
	err := db.Set(key, value)
	if err != nil {
		db.log.Error("%v", err)
		panic(err)
	}
}

// mustSet writes a key-value pair without a lock and panics on error.
func (db *LogDB) mustSet(key string, value []byte) {
	err := db.set(key, value)
	if err != nil {
		db.log.Error("%v", err)
		panic(err)
	}
}

// Aggregate gathers all the unique key-value pairs in the database,
// and then removes all the segments and resets the state.
func (db *LogDB) Aggregate() map[string][]byte {
	db.log.Info("Aggregating segments")
	db.Lock()
	defer db.Unlock()

	values := make(map[string][]byte, len(db.head.hashIndex))
	current := db.head
	for {
		for key := range current.hashIndex {
			if _, ok := values[key]; !ok {
				value, _ := current.get(key)
				values[key] = value
			}
		}

		// Check if we've reached the end of the linked list.
		current.Lock()
		if current.next == db.head || current.next == nil {
			err := current.delete()
			if err != nil {
				db.log.Error(err)
			}
			current.next = nil
			current.Unlock()
			break
		}

		err := current.delete()
		if err != nil {
			db.log.Error(err)
		}

		current.prev.next = nil
		current.prev = nil
		current.Unlock()
		current = current.next
	}

	db.head = nil
	db.tail = nil

	segment := newSegment(db.dirPath, 0)
	db.head, db.tail = segment, nil

	db.log.Info("Aggregation completed")
	return values
}
