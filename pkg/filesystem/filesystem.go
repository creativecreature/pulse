package filesystem

type GitFile interface {
	Name() string
	Filetype() string
	Repository() string
	Path() string
}
