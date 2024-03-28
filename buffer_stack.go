package codeharvest

// bufferStack represents the buffers that have been opened during a coding session.
type bufferStack struct {
	buffers []Buffer
}

// newBufferStack creates a new buffer stack.
func newBufferStack() *bufferStack {
	buffers := make([]Buffer, 0)
	return &bufferStack{buffers}
}

// push pushes a buffer onto the stack.
func (s *bufferStack) push(f Buffer) {
	s.buffers = append(s.buffers, f)
}

// peek returns a pointer to the most recent buffer.
func (s *bufferStack) peek() *Buffer {
	if len(s.buffers) == 0 {
		return nil
	}
	return &s.buffers[len(s.buffers)-1]
}

// files takes the stack of buffers, merges them by filepath,
// and returns the result in the order they were opened in.
func (s *bufferStack) files() Files {
	sortOrder := make([]string, 0)
	pathFile := make(map[string]File)

	// Merge the buffers by filepath.
	for _, buffer := range s.buffers {
		if file, exists := pathFile[buffer.Filepath]; !exists {
			sortOrder = append(sortOrder, buffer.Filepath)
			pathFile[buffer.Filepath] = fileFromBuffer(buffer)
		} else {
			file.DurationMs += buffer.Duration()
			pathFile[buffer.Filepath] = file
		}
	}

	// Return the buffers in the original order.
	files := make(Files, 0, len(pathFile))
	for _, path := range sortOrder {
		files = append(files, pathFile[path])
	}

	return files
}
