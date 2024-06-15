package logdb

import "strings"

// Filename generates a filename based on the given index.
func Filename(index int) string {
	length := 16
	result := make([]byte, length)
	copy(result, strings.Repeat("a", length))

	for pos := length - 1; index > 0 && pos >= 0; pos-- {
		result[pos] = 'a' + byte(index%26)
		index /= 26
	}

	return string(result) + ".log"
}

// Index extracts an index from a filename.
func Index(filename string) int {
	sequence := strings.TrimSuffix(filename, ".log")
	index, multiplier := 0, 1
	for pos := len(sequence) - 1; pos >= 0; pos-- {
		charIndex := sequence[pos] - 'a'
		index += int(charIndex) * multiplier
		multiplier *= 26
	}
	return index
}
