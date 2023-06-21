package filereader_test

import (
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"code-harvest.conner.dev/filereader"
)

type fileReaderMock struct {
	directoryIndex int
	directories    []string
	entries        map[string][]fs.DirEntry
	fileContents   map[string][]byte
	filereader.Reader
}

func (f *fileReaderMock) Dir(_ string) string {
	if f.directoryIndex > len(f.directories)-1 {
		return ""
	}
	dir := f.directories[f.directoryIndex]
	f.directoryIndex++
	return dir
}

func (f *fileReaderMock) ReadDir(dir string) ([]fs.DirEntry, error) {
	entries, ok := f.entries[dir]
	if !ok {
		return nil, fmt.Errorf("no entries for dir: %s", dir)
	}
	return entries, nil
}

func (f *fileReaderMock) ReadFile(filename string) ([]byte, error) {
	fileContent, ok := f.fileContents[filename]
	if !ok {
		return nil, fmt.Errorf("no file content for file: %s", filename)
	}
	return fileContent, nil
}

func (f *fileReaderMock) IsFile(filename string) bool {
	return true
}

func (f *fileReaderMock) Filename(path string) string {
	s := strings.Split(path, "/")
	return s[len(s)-1]
}

type fileEntryMock struct {
	fs.DirEntry
	Filename    string
	IsDirectory bool
}

func (f fileEntryMock) Name() string {
	return f.Filename
}

func (f fileEntryMock) IsDir() bool {
	return f.IsDirectory
}

func TestGetRepositoryFromPath(t *testing.T) {
	t.Parallel()

	// The config file that we expect to find within the git directory
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
			fileEntryMock{Filename: "init.lua", IsDirectory: false},
			fileEntryMock{Filename: "keybindings.lua", IsDirectory: false},
			fileEntryMock{Filename: "autocommands.lua", IsDirectory: false},
		},
		"/Users/conner/code/dotfiles/editors": {
			fileEntryMock{Filename: "nvim", IsDirectory: true},
		},
		"/Users/conner/code/dotfiles": {
			fileEntryMock{Filename: "editors", IsDirectory: true},
			fileEntryMock{Filename: "bootstrap.sh", IsDirectory: false},
			fileEntryMock{Filename: "install.sh", IsDirectory: false},
			fileEntryMock{Filename: ".git", IsDirectory: true},
		},
	}

	fileSystemMock := fileReaderMock{
		directoryIndex: 0,
		directories: []string{
			"/Users/conner/code/dotfiles/editors/nvim",
			"/Users/conner/code/dotfiles/editors",
			"/Users/conner/code/dotfiles",
			"/Users/conner/code",
			"/Users/conner",
			"/Users",
			"/",
		},
		entries: directoryEntries,
		fileContents: map[string][]byte{
			"/Users/conner/code/dotfiles/.git/config": []byte(gitConfigFile),
		},
	}

	f := filereader.New()
	f.Reader = &fileSystemMock

	// This is the absolute path of the file that we want to extract the repository name for.
	path := "/Users/conner/code/dotfiles/editors/nvim/init.lua"
	file, err := f.GitFile(path)
	if err != nil {
		t.Fatal(err)
	}

	// From how the mocks are wired we expect dotfiles to be the repository name.
	expected := "dotfiles"
	got := file.Repository()

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
			fileEntryMock{Filename: "index.ts", IsDirectory: false},
			fileEntryMock{Filename: "index.html", IsDirectory: false},
			fileEntryMock{Filename: "components", IsDirectory: true},
		},
		"/Users/conner/code/ore-ui/main": {
			fileEntryMock{Filename: "src", IsDirectory: true},
			fileEntryMock{Filename: ".git", IsDirectory: false},
		},
		"/Users/conner/code/ore-ui": {
			fileEntryMock{Filename: "main", IsDirectory: true},
			fileEntryMock{Filename: "dev", IsDirectory: true},
			fileEntryMock{Filename: ".bare", IsDirectory: true},
			fileEntryMock{Filename: ".git", IsDirectory: false},
		},
	}

	fileSystemMock := fileReaderMock{
		directoryIndex: 0,
		directories: []string{
			"/Users/conner/code/ore-ui/main/src",
			"/Users/conner/code/ore-ui/main",
			"/Users/conner/code/ore-ui",
			"/Users/conner/code",
			"/Users/conner",
			"/Users",
			"/",
		},
		entries: directoryEntries,
		fileContents: map[string][]byte{
			"/Users/conner/code/ore-ui/main/.git":    []byte(gitFile),
			"/Users/conner/code/ore-ui/.bare/config": []byte(gitConfigFile),
		},
	}

	f := filereader.New()
	f.Reader = &fileSystemMock

	// This is the absolute path of the file that we want to extract the repository name for.
	path := "/Users/conner/code/ore-ui/main/src/index.ts"
	file, err := f.GitFile(path)
	if err != nil {
		t.Fatal(err)
	}

	// From how the mocks are wired we expect ore-ui to be the repository name.
	expected := "ore-ui"
	got := file.Repository()

	if got != expected {
		t.Errorf("GetRepositoryFromPath(%s) = %s; expected %s", path, got, expected)
	}
}

func TestPathInBareProject(t *testing.T) {
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
			fileEntryMock{Filename: "index.ts", IsDirectory: false},
			fileEntryMock{Filename: "index.html", IsDirectory: false},
			fileEntryMock{Filename: "components", IsDirectory: true},
		},
		"/Users/conner/code/ore-ui/main": {
			fileEntryMock{Filename: "src", IsDirectory: true},
			fileEntryMock{Filename: ".git", IsDirectory: false},
		},
		"/Users/conner/code/ore-ui": {
			fileEntryMock{Filename: "main", IsDirectory: true},
			fileEntryMock{Filename: "dev", IsDirectory: true},
			fileEntryMock{Filename: ".bare", IsDirectory: true},
			fileEntryMock{Filename: ".git", IsDirectory: false},
		},
	}

	fileSystemMock := fileReaderMock{
		directoryIndex: 0,
		directories: []string{
			"/Users/conner/code/ore-ui/main/src",
			"/Users/conner/code/ore-ui/main",
			"/Users/conner/code/ore-ui",
			"/Users/conner/code",
			"/Users/conner",
			"/Users",
			"/",
		},
		entries: directoryEntries,
		fileContents: map[string][]byte{
			"/Users/conner/code/ore-ui/main/.git":    []byte(gitFile),
			"/Users/conner/code/ore-ui/.bare/config": []byte(gitConfigFile),
		},
	}

	f := filereader.New()
	f.Reader = &fileSystemMock

	// This is the absolute path of the file that we want to extract the repository name for.
	path := "/Users/conner/code/ore-ui/main/src/index.ts"
	file, err := f.GitFile(path)
	if err != nil {
		t.Fatal(err)
	}

	// From how the mocks are wired we expect ore-ui to be the repository name.
	expected := "ore-ui"
	got := file.Repository()

	if got != expected {
		t.Errorf("GetRepositoryFromPath(%s) = %s; expected %s", path, got, expected)
	}
}

func TestPathInProject(t *testing.T) {
	t.Parallel()

	// The config file that we expect to find within the git directory
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
			fileEntryMock{Filename: "init.lua", IsDirectory: false},
			fileEntryMock{Filename: "keybindings.lua", IsDirectory: false},
			fileEntryMock{Filename: "autocommands.lua", IsDirectory: false},
		},
		"/Users/conner/code/dotfiles/editors": {
			fileEntryMock{Filename: "nvim", IsDirectory: true},
		},
		"/Users/conner/code/dotfiles": {
			fileEntryMock{Filename: "editors", IsDirectory: true},
			fileEntryMock{Filename: "bootstrap.sh", IsDirectory: false},
			fileEntryMock{Filename: "install.sh", IsDirectory: false},
			fileEntryMock{Filename: ".git", IsDirectory: true},
		},
	}

	fileSystemMock := fileReaderMock{
		directoryIndex: 0,
		directories: []string{
			"/Users/conner/code/dotfiles/editors/nvim",
			"/Users/conner/code/dotfiles/editors",
			"/Users/conner/code/dotfiles",
			"/Users/conner/code",
			"/Users/conner",
			"/Users",
			"/",
		},
		entries: directoryEntries,
		fileContents: map[string][]byte{
			"/Users/conner/code/dotfiles/.git/config": []byte(gitConfigFile),
		},
	}

	f := filereader.New()
	f.Reader = &fileSystemMock

	// This is the absolute path of the file that we want to extract the repository name for.
	path := "/Users/conner/code/dotfiles/editors/nvim/init.lua"
	file, err := f.GitFile(path)
	if err != nil {
		t.Fatal(err)
	}

	// From how the mocks are wired we expect dotfiles to be the repository name.
	expected := "dotfiles/editors/nvim/init.lua"
	got := file.Path()

	if got != expected {
		t.Errorf("GetRepositoryFromPath(%s) = %s; expected %s", path, got, expected)
	}
}
