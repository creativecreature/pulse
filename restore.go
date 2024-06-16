package pulse

import (
	"os"
	"path"
	"path/filepath"
	"sort"
)

// getSegmentPaths returns a sorted list of every segments log file in the directory.
func getSegmentPaths(dirPath string) ([]string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	filePaths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filePaths = append(filePaths, path.Join(dirPath, entry.Name()))
	}

	sort.Slice(filePaths, func(i, j int) bool {
		return filePaths[i] > filePaths[j]
	})
	return filePaths, nil
}

// restoreSegment reads a log file and restores it to a segment.
func restoreSegment(path string) (*Segment, error) {
	var bytes int64
	hashIndex := make(HashIndex)
	for record := range scan(path) {
		hashIndex[record.Key] = record.Offset
		bytes += int64(len(record.Key)) + int64(len(record.Value))
	}

	file, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, err
	}

	filename := filepath.Base(path)
	segment := &Segment{
		index:     Index(filename),
		bytes:     bytes,
		hashIndex: hashIndex,
		logFile:   file,
	}

	return segment, nil
}

// connectSegments links all segments together in a circular doubly linked list.
func connectSegments(segments []*Segment) {
	for i := 0; i < len(segments); i++ {
		if i == 0 {
			segments[i].prev = segments[len(segments)-1]
			segments[i].next = segments[i+1]
			continue
		}

		if i == len(segments)-1 {
			segments[i].prev = segments[i-1]
			segments[i].next = segments[0]
			continue
		}

		segments[i].prev = segments[i-1]
		segments[i].next = segments[i+1]
	}
}

// restoreSegments reads all log files in the directory and restores them to segments.
func restoreSegments(segmentPaths []string) []*Segment {
	segments := make([]*Segment, 0, len(segmentPaths))
	for _, p := range segmentPaths {
		segment, err := restoreSegment(p)
		if err != nil {
			panic(err)
		}
		segments = append(segments, segment)
	}

	if len(segments) > 1 {
		connectSegments(segments)
	}

	return segments
}
