package git

import (
	"io/fs"
	"testing"

	"code-harvest.conner.dev/pkg/filesystem"
)

func TestGetRepositoryFromPath(t *testing.T) {
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
			filesystem.MockFileEntry{Filename: "init.lua", IsDirectory: false},
			filesystem.MockFileEntry{Filename: "keybindings.lua", IsDirectory: false},
			filesystem.MockFileEntry{Filename: "autocommands.lua", IsDirectory: false},
		},
		"/Users/conner/code/dotfiles/editors": {
			filesystem.MockFileEntry{Filename: "nvim", IsDirectory: true},
		},
		"/Users/conner/code/dotfiles": {
			filesystem.MockFileEntry{Filename: "editors", IsDirectory: true},
			filesystem.MockFileEntry{Filename: "bootstrap.sh", IsDirectory: false},
			filesystem.MockFileEntry{Filename: "install.sh", IsDirectory: false},
			filesystem.MockFileEntry{Filename: ".git", IsDirectory: true},
		},
	}

	fileSystemMock := filesystem.MockFS{
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
	got, _ := GetRepositoryFromPath(&fileSystemMock, path)
	// From how the mocks are wired we expect dotfiles to be the repository name.
	expected := "dotfiles"

	if got != expected {
		t.Errorf("GetRepositoryFromPath(%s) = %s; expected %s", path, got, expected)
	}
}

func TestGetRepositoryFromPathBare(t *testing.T) {
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
			filesystem.MockFileEntry{Filename: "index.ts", IsDirectory: false},
			filesystem.MockFileEntry{Filename: "index.html", IsDirectory: false},
			filesystem.MockFileEntry{Filename: "components", IsDirectory: true},
		},
		"/Users/conner/code/ore-ui/main": {
			filesystem.MockFileEntry{Filename: "src", IsDirectory: true},
			filesystem.MockFileEntry{Filename: ".git", IsDirectory: false},
		},
		"/Users/conner/code/ore-ui": {
			filesystem.MockFileEntry{Filename: "main", IsDirectory: true},
			filesystem.MockFileEntry{Filename: "dev", IsDirectory: true},
			filesystem.MockFileEntry{Filename: ".bare", IsDirectory: true},
			filesystem.MockFileEntry{Filename: ".git", IsDirectory: false},
		},
	}

	fileSystemMock := filesystem.MockFS{
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

	// Monkey patch the functions that hit the OS/Filesystem with our mocks.

	// This is the absolute path of the file that we want to extract the repository name for.
	path := "/Users/conner/code/ore-ui/main/src/index.ts"
	got, _ := GetRepositoryFromPath(&fileSystemMock, path)
	// From how the mocks are wired we expect ore-ui to be the repository name.
	expected := "ore-ui"

	if got != expected {
		t.Errorf("GetRepositoryFromPath(%s) = %s; expected %s", path, got, expected)
	}
}

func TestGetRelativePathFromRepo(t *testing.T) {
	tests := []struct {
		path       string
		repository string
		expected   string
	}{
		{"/Users/conner/code/project/src/index.html", "project", "project/src/index.html"},
		{"/Users/conner/code/dotfiles/editors/nvim/init.lua", "dotfiles", "dotfiles/editors/nvim/init.lua"},
		{"/Users/conner/code/dotfiles/editors/nvim/init.lua", "dotfiles", "dotfiles/editors/nvim/init.lua"},
	}

	for _, test := range tests {
		got, _ := GetRelativePathFromRepo(test.path, test.repository)
		if got != test.expected {
			t.Errorf("GetRelativePathFromRepo(%s, %s) = %s; wanted %s;", test.path, test.repository, got, test.expected)
		}
	}
}
