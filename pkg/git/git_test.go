package git_test

import (
	"errors"
	"io/fs"
	"testing"

	"code-harvest.conner.dev/pkg/git"
)

type MockFS struct {
	DirectoryIndex int
	Directories    []string
	Entries        map[string][]fs.DirEntry
	FileContents   map[string][]byte
	git.Filesystem
}

func (f *MockFS) Dir(_ string) string {
	if f.DirectoryIndex > len(f.Directories)-1 {
		return ""
	}
	dir := f.Directories[f.DirectoryIndex]
	f.DirectoryIndex++
	return dir
}

func (f *MockFS) ReadDir(dir string) ([]fs.DirEntry, error) {
	entries, ok := f.Entries[dir]
	if !ok {
		return nil, errors.New("no entries for dir")
	}
	return entries, nil
}

func (f *MockFS) ReadFile(filename string) ([]byte, error) {
	fileContent, ok := f.FileContents[filename]
	if !ok {
		return nil, errors.New("no content for this filename")
	}
	return fileContent, nil
}

type MockFileEntry struct {
	fs.DirEntry
	Filename    string
	IsDirectory bool
}

func (f MockFileEntry) Name() string {
	return f.Filename
}

func (f MockFileEntry) IsDir() bool {
	return f.IsDirectory
}

func TestGetRepositoryFromPath(t *testing.T) {
	t.Parallel()

	// The config file that we expect to find within the git directory
	// OCD vs Correctness. Should obviously not be indented.
	gitConfigFile := `
		[core]
			repositoryformatversion = 0
			filemode = true
			bare = false
			logallrefupdates = true
			ignorecase = true
			precomposeunicode = true
		[remote "origin"]
			url = git@github.com:creativecreature/dotfiles.git
			fetch = +refs/heads/*:refs/remotes/origin/*
			gh-resolved = base
		[branch "master"]
			remote = origin
			merge = refs/heads/master
	`

	// Set up the entries we expect to see for each directory.
	directoryEntries := map[string][]fs.DirEntry{
		"/Users/conner/code/dotfiles/editors/nvim": {
			MockFileEntry{Filename: "init.lua", IsDirectory: false},
			MockFileEntry{Filename: "keybindings.lua", IsDirectory: false},
			MockFileEntry{Filename: "autocommands.lua", IsDirectory: false},
		},
		"/Users/conner/code/dotfiles/editors": {
			MockFileEntry{Filename: "nvim", IsDirectory: true},
		},
		"/Users/conner/code/dotfiles": {
			MockFileEntry{Filename: "editors", IsDirectory: true},
			MockFileEntry{Filename: "bootstrap.sh", IsDirectory: false},
			MockFileEntry{Filename: "install.sh", IsDirectory: false},
			MockFileEntry{Filename: ".git", IsDirectory: true},
		},
	}

	fileSystemMock := MockFS{
		DirectoryIndex: 0,
		Directories: []string{
			"/Users/conner/code/dotfiles/editors/nvim",
			"/Users/conner/code/dotfiles/editors",
			"/Users/conner/code/dotfiles",
			"/Users/conner/code",
			"/Users/conner",
			"/Users",
			"/",
		},
		Entries: directoryEntries,
		FileContents: map[string][]byte{
			"/Users/conner/code/dotfiles/.git/config": []byte(gitConfigFile),
		},
	}

	// This is the absolute path of the file that we want to extract the repository name for.
	path := "/Users/conner/code/dotfiles/editors/nvim/init.lua"
	g := git.New(&fileSystemMock)
	got, _ := g.RepositoryName(path)
	// From how the mocks are wired we expect dotfiles to be the repository name.
	expected := "dotfiles"

	if got != expected {
		t.Errorf("GetRepositoryFromPath(%s) = %s; expected %s", path, got, expected)
	}
}

func TestGetRepositoryFromPathBare(t *testing.T) {
	t.Parallel()

	// When I use git worktrees I make .git a file that points to the location of
	// the bare directory. It's important to note that each worktree has its own .gitfile
	gitFile := `gitdir: /Users/conner/code/ore-ui/.bare/worktrees/main`
	gitConfigFile := `
		[core]
			repositoryformatversion = 0
			filemode = true
			bare = true
			ignorecase = true
			precomposeunicode = true
		[remote "origin"]
			url = git@github.com:Mojang/ore-ui.git
			fetch = +refs/heads/*:refs/remotes/origin/*
		[branch "main"]
			remote = origin
			merge = refs/heads/main
	`

	// Set up the entries we expect to see for each directory.
	directoryEntries := map[string][]fs.DirEntry{
		"/Users/conner/code/ore-ui/main/src": {
			MockFileEntry{Filename: "index.ts", IsDirectory: false},
			MockFileEntry{Filename: "index.html", IsDirectory: false},
			MockFileEntry{Filename: "components", IsDirectory: true},
		},
		"/Users/conner/code/ore-ui/main": {
			MockFileEntry{Filename: "src", IsDirectory: true},
			MockFileEntry{Filename: ".git", IsDirectory: false},
		},
		"/Users/conner/code/ore-ui": {
			MockFileEntry{Filename: "main", IsDirectory: true},
			MockFileEntry{Filename: "dev", IsDirectory: true},
			MockFileEntry{Filename: ".bare", IsDirectory: true},
			MockFileEntry{Filename: ".git", IsDirectory: false},
		},
	}

	fileSystemMock := MockFS{
		DirectoryIndex: 0,
		Directories: []string{
			"/Users/conner/code/ore-ui/main/src",
			"/Users/conner/code/ore-ui/main",
			"/Users/conner/code/ore-ui",
			"/Users/conner/code",
			"/Users/conner",
			"/Users",
			"/",
		},
		Entries: directoryEntries,
		FileContents: map[string][]byte{
			"/Users/conner/code/ore-ui/main/.git":    []byte(gitFile),
			"/Users/conner/code/ore-ui/.bare/config": []byte(gitConfigFile),
		},
	}

	// This is the absolute path of the file that we want to extract the repository name for.
	path := "/Users/conner/code/ore-ui/main/src/index.ts"
	g := git.New(&fileSystemMock)
	got, _ := g.RepositoryName(path)
	// From how the mocks are wired we expect ore-ui to be the repository name.
	expected := "ore-ui"

	if got != expected {
		t.Errorf("GetRepositoryFromPath(%s) = %s; expected %s", path, got, expected)
	}
}
