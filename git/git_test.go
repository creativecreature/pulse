package git_test

import (
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"github.com/viccon/pulse/git"
)

type readerMock struct {
	DirectoryIndex int
	Directories    []string
	Entries        map[string][]fs.DirEntry
	FileContents   map[string][]byte
}

func (f *readerMock) Dir(_ string) string {
	if f.DirectoryIndex > len(f.Directories)-1 {
		return ""
	}
	dir := f.Directories[f.DirectoryIndex]
	f.DirectoryIndex++
	return dir
}

func (f *readerMock) ReadDir(dir string) ([]fs.DirEntry, error) {
	entries, ok := f.Entries[dir]
	if !ok {
		return nil, fmt.Errorf("no entries for dir: %s", dir)
	}
	return entries, nil
}

func (f *readerMock) ReadFile(filename string) ([]byte, error) {
	fileContent, ok := f.FileContents[filename]
	if !ok {
		return nil, fmt.Errorf("no file content for file: %s", filename)
	}
	return fileContent, nil
}

func (f *readerMock) IsFile(_ string) bool {
	return true
}

func (f *readerMock) Filename(path string) string {
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

func newFileEntry(filename string, isDirectory bool) fileEntryMock {
	return fileEntryMock{
		Filename:    filename,
		IsDirectory: isDirectory,
	}
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
			url = git@github.com:viccon/dotfiles.git
			fetch = +refs/heads/*:refs/remotes/origin/*
			gh-resolved = base
		[branch "master"]
			remote = origin
			merge = refs/heads/master
	`

	// Set up the entries we expect to see for each directory.
	directoryEntries := map[string][]fs.DirEntry{
		"/Users/conner/code/dotfiles/editors/nvim": {
			newFileEntry("init.lua", false),
			newFileEntry("keybindings.lua", false),
			newFileEntry("autocommands.lua", false),
		},
		"/Users/conner/code/dotfiles/editors": {
			newFileEntry("nvim", true),
		},
		"/Users/conner/code/dotfiles": {
			newFileEntry("editors", true),
			newFileEntry("bootstrap.sh", false),
			newFileEntry("install.sh", false),
			newFileEntry(".git", true),
		},
	}

	fileSystemMock := readerMock{
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

	f := git.New()
	f.Reader = &fileSystemMock

	// This is the absolute path of the file that we want to extract the repository name for.
	path := "/Users/conner/code/dotfiles/editors/nvim/init.lua"
	file, err := f.ParseFile(path, "lua")
	if err != nil {
		t.Fatal(err)
	}

	// From how the mocks are wired we expect dotfiles to be the repository name.
	expected := "dotfiles"
	got := file.Repository

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
			newFileEntry("index.ts", false),
			newFileEntry("index.html", false),
			newFileEntry("components", true),
		},
		"/Users/conner/code/ore-ui/main": {
			newFileEntry("src", true),
			newFileEntry(".git", false),
		},
		"/Users/conner/code/ore-ui": {
			newFileEntry("main", true),
			newFileEntry("dev", true),
			newFileEntry(".bare", true),
			newFileEntry(".git", false),
		},
	}

	fileSystemMock := readerMock{
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

	f := git.New()
	f.Reader = &fileSystemMock

	// This is the absolute path of the file that we want to extract the repository name for.
	path := "/Users/conner/code/ore-ui/main/src/index.ts"
	file, err := f.ParseFile(path, "typescript")
	if err != nil {
		t.Fatal(err)
	}

	// From how the mocks are wired we expect ore-ui to be the repository name.
	expected := "ore-ui"
	got := file.Repository

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
			newFileEntry("index.ts", false),
			newFileEntry("index.html", false),
			newFileEntry("components", true),
		},
		"/Users/conner/code/ore-ui/main": {
			newFileEntry("src", true),
			newFileEntry(".git", false),
		},
		"/Users/conner/code/ore-ui": {
			newFileEntry("main", true),
			newFileEntry("dev", true),
			newFileEntry(".bare", true),
			newFileEntry(".git", false),
		},
	}

	fileSystemMock := readerMock{
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

	f := git.New()
	f.Reader = &fileSystemMock

	// This is the absolute path of the file that we want to extract the repository name for.
	path := "/Users/conner/code/ore-ui/main/src/index.ts"
	file, err := f.ParseFile(path, "typescript")
	if err != nil {
		t.Fatal(err)
	}

	// From how the mocks are wired we expect ore-ui to be the repository name.
	expected := "ore-ui"
	got := file.Repository

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
			url = git@github.com:viccon/dotfiles.git
			fetch = +refs/heads/*:refs/remotes/origin/*
			gh-resolved = base
		[branch "master"]
			remote = origin
			merge = refs/heads/master
	`

	// Set up the entries we expect to see for each directory.
	directoryEntries := map[string][]fs.DirEntry{
		"/Users/conner/code/dotfiles/editors/nvim": {
			newFileEntry("init.lua", false),
			newFileEntry("keybindings.lua", false),
			newFileEntry("autocommands.lua", false),
		},
		"/Users/conner/code/dotfiles/editors": {
			newFileEntry("nvim", true),
		},
		"/Users/conner/code/dotfiles": {
			newFileEntry("editors", true),
			newFileEntry("bootstrap.sh", false),
			newFileEntry("install.sh", false),
			newFileEntry(".git", true),
		},
	}

	fileSystemMock := readerMock{
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

	f := git.New()
	f.Reader = &fileSystemMock

	// This is the absolute path of the file that we want to extract the repository name for.
	path := "/Users/conner/code/dotfiles/editors/nvim/init.lua"
	file, err := f.ParseFile(path, "typescript")
	if err != nil {
		t.Fatal(err)
	}

	// From how the mocks are wired we expect dotfiles to be the repository name.
	expected := "dotfiles/editors/nvim/init.lua"
	got := file.Path

	if got != expected {
		t.Errorf("GetRepositoryFromPath(%s) = %s; expected %s", path, got, expected)
	}
}
