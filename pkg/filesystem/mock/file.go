package mock

type File struct {
	name       string
	filetype   string
	repository string
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

func NewFile(name, filetype, repository string) File {
	return File{name, filetype, repository}
}
