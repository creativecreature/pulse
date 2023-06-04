package filesystem

type File interface {
	Name() string
	Filetype() string
	Repository() string
	Path() string
}

type file struct {
	name       string
	filetype   string
	repository string
	path       string
}

func (f file) Name() string {
	return f.name
}

func (f file) Filetype() string {
	return f.filetype
}

func (f file) Repository() string {
	return f.repository
}

func (f file) Path() string {
	return f.path
}
