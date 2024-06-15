package logdb

import (
	"encoding/json"
	"io"
	"os"
	"path"
	"sync"
)

// HashIndex is a map of keys to offsets in the segment file.
type HashIndex map[string]int64

// Segment represents a segment in our log database. Each
// segment has its own file descriptor and hash index.
type Segment struct {
	sync.RWMutex
	index     int
	bytes     int64
	prev      *Segment
	next      *Segment
	hashIndex HashIndex
	logFile   *os.File
}

// newSegment creates a new segment with the given index.
func newSegment(dirpath string, segmentIndex int) *Segment {
	fileName := Filename(segmentIndex)
	file, err := os.Create(path.Join(dirpath, fileName))
	if err != nil {
		panic(err)
	}

	newSegment := &Segment{
		index:     segmentIndex,
		hashIndex: make(HashIndex),
		logFile:   file,
	}

	return newSegment
}

// get retrieves a value from the segment.
func (s *Segment) get(key string) ([]byte, bool) {
	s.Lock()
	defer s.Unlock()

	offset, ok := s.hashIndex[key]
	if !ok {
		return nil, false
	}
	_, err := s.logFile.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, false
	}

	var record Record
	decoder := json.NewDecoder(s.logFile)
	if decodeErr := decoder.Decode(&record); decodeErr != nil {
		return nil, false
	}

	return record.Value, true
}

// set writes a key-value pair to the segments log file.
func (s *Segment) set(key string, value []byte) error {
	s.Lock()
	defer s.Unlock()

	offset, err := s.logFile.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	r := Record{Key: key, Value: value}
	bytes, err := json.Marshal(r)
	if err != nil {
		return err
	}

	bytes = append(bytes, '\n')
	_, err = s.logFile.Write(bytes)
	if err != nil {
		return err
	}
	s.hashIndex[key] = offset
	s.bytes = offset + int64(len(bytes))

	return nil
}

// size returns the size of the segment in bytes.
func (s *Segment) size() int64 {
	s.Lock()
	defer s.Unlock()
	return s.bytes
}

// delete closes the file descriptor and removes the segment file from disk.
func (s *Segment) delete() {
	s.Lock()
	defer s.Unlock()

	s.logFile.Close()
	err := os.Remove(s.logFile.Name())
	if err != nil {
		panic(err)
	}
}
