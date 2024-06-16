package pulse

import (
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

const (
	segmentationInterval = 5 * time.Minute
	segmentSizeBytes     = 10 * 1024 // 10KB
)

// Record represents a key-value pair in our database.
type Record struct {
	Key   string `json:"key"`
	Value []byte `json:"value"`
}

// LogDB is a simple key-value store that persists data to a log file.
type LogDB struct {
	sync.RWMutex
	dirPath string
	log     *log.Logger
	head    *Segment
	tail    *Segment
}

// NewDB creates a new log database.
func NewDB(dirPath string) *LogDB {
	log := NewLogger()

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
	logDB.log = log

	// Leak a goroutine that compacts the segments.
	defer func() {
		go func() {
			ticker := time.NewTicker(segmentationInterval)
			for {
				<-ticker.C
				logDB.compact()
			}
		}()
	}()

	// If the directory is empty, we'll simply create the initial segment and return.
	if len(segmentPaths) == 0 {
		segment := newSegment(dirPath, 0)
		logDB.head, logDB.tail = segment, segment
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

// appendSegment creates a new segment and appends it to the
// head of the linked list. should be called with a lock.
func (db *LogDB) appendSegment() {
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
	db.log.Info("Compacting segments")

	head, current := db.head, db.tail
	if current == nil {
		db.log.Info("No segments to compact")
		return
	}

	for {
		for key := range current.hashIndex {
			var found bool

			for cursor := head; cursor != current; cursor = cursor.next {
				_, found = cursor.get(key)
				if found {
					current.Lock()
					delete(current.hashIndex, key)
					current.Unlock()
					break
				}
			}

			// If this key was unique for all previous segments, we'll write it to the head.
			if !found {
				bytes, _ := current.get(key)
				db.MustSet(key, bytes)
			}
		}

		// Delete the segment file once we've compacted it.
		current.Lock()
		current.delete()
		current.Unlock()

		// Exit the loop if we've reached the head.
		if current.prev == head {
			break
		}

		db.Lock()
		current = current.prev
		current.next = db.head
		db.head.prev = current
		db.tail = current
		db.Unlock()
	}
	log.Info("Finished compacting segments")
}

// Get retrieves a value from the database.
func (db *LogDB) Get(key string) ([]byte, bool) {
	db.log.Debug("Getting key", key)
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

// Set writes a key-value pair to the log file.
func (db *LogDB) Set(key string, value []byte) error {
	db.log.Debug("writing key", key)
	db.Lock()
	defer db.Unlock()

	err := db.head.set(key, value)
	if err != nil {
		return err
	}
	if db.head.size() >= segmentSizeBytes {
		db.appendSegment()
	}
	return nil
}

// MustSet writes a key-value pair to the log file and panics on error.
func (db *LogDB) MustSet(key string, value []byte) {
	err := db.Set(key, value)
	if err != nil {
		panic(err)
	}
}

// Aggregate gathers all the unique key-value pairs in the database,
// and then removes all the segments and resets the state.
func (db *LogDB) Aggregate() map[string][]byte {
	db.log.Debug("Aggregating segments")
	db.Lock()
	defer db.Unlock()

	values := make(map[string][]byte, len(db.head.hashIndex))

	// Break the connection between the tail and head.
	if db.tail != nil {
		db.tail.next = nil
	}

	for current := db.head; current != nil; {
		for key := range current.hashIndex {
			if _, ok := values[key]; !ok {
				value, _ := current.get(key)
				values[key] = value
			}
			delete(current.hashIndex, key)
		}

		current.Lock()
		current.delete()
		current.Unlock()

		// Update current and break if we've reached the tail.
		current = current.next
		if current == nil {
			break
		}

		// Update the references so it can be GC'd.
		current.prev.next = nil
		current.prev = nil
	}

	segment := newSegment(db.dirPath, 0)
	db.head, db.tail = segment, segment

	db.log.Debug("Finished the aggrementation process")
	return values
}
