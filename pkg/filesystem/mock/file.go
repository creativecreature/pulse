package mock

type File struct {
	name       string
	filetype   string
	repository string
	path       string
}

func (m File) Name() string {
	return m.name
}

func (m File) Filetype() string {
	return m.filetype
}

func (m File) Repository() string {
	return m.repository
}

func (m File) Path() string {
	return m.path
}

func NewFile(name, filetype, repository, path string) File {
	return File{name, filetype, repository, path}
}
