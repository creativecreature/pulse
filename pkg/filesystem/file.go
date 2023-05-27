package filesystem

type File interface {
	Name() string
	Filetype() string
	Repository() string
}

type file struct {
	name       string
	filetype   string
	repository string
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
