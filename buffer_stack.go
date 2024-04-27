package pulse

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
func (b *bufferStack) push(f Buffer) {
	b.buffers = append(b.buffers, f)
}

// peek returns a pointer to the most recent buffer.
func (b *bufferStack) peek() *Buffer {
	if len(b.buffers) == 0 {
		return nil
	}
	return &b.buffers[len(b.buffers)-1]
}

// files takes the stack of buffers, merges them by filepath,
// and returns the result in the order they were opened in.
func (b *bufferStack) files() Files {
	sortOrder := make([]string, 0)
	pathFile := make(map[string]File)

	// Merge the buffers by filepath.
	for _, buffer := range b.buffers {
		file, exists := pathFile[buffer.Filepath]
		if !exists {
			sortOrder = append(sortOrder, buffer.Filepath)
			pathFile[buffer.Filepath] = fileFromBuffer(buffer)
			continue
		}
		file.Duration += buffer.Duration()
		pathFile[buffer.Filepath] = file
	}

	// Return the buffers in the original order.
	files := make(Files, 0, len(pathFile))
	for _, path := range sortOrder {
		files = append(files, pathFile[path])
	}

	return files
}
