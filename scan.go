package pulse

import (
	"bufio"
	"encoding/json"
	"os"
)

// RecordWithOffset holds a record and its offset in the log file.
type RecordWithOffset struct {
	Record
	Offset int64
}

// scan reads a log file and sends each record to a channel along with its offset.
func scan(filepath string) <-chan RecordWithOffset {
	ch := make(chan RecordWithOffset)

	file, err := os.Open(filepath)
	if err != nil {
		close(ch)
		return ch
	}

	go func() {
		defer close(ch)
		defer file.Close()

		var currentOffset int64
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			var record Record
			if unmarshalErr := json.Unmarshal(scanner.Bytes(), &record); unmarshalErr != nil {
				continue
			}
			ch <- RecordWithOffset{record, currentOffset}
			currentOffset += int64(len(scanner.Bytes()) + len("\n"))
		}
	}()

	return ch
}
