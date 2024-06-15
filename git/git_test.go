package git_test

import (
	"io/fs"
	"testing"

	"github.com/creativecreature/pulse/git"
	"github.com/creativecreature/pulse/mock"
)

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
			mock.NewFileEntry("init.lua", false),
			mock.NewFileEntry("keybindings.lua", false),
			mock.NewFileEntry("autocommands.lua", false),
		},
		"/Users/conner/code/dotfiles/editors": {
			mock.NewFileEntry("nvim", true),
		},
		"/Users/conner/code/dotfiles": {
			mock.NewFileEntry("editors", true),
			mock.NewFileEntry("bootstrap.sh", false),
			mock.NewFileEntry("install.sh", false),
			mock.NewFileEntry(".git", true),
		},
	}

	fileSystemMock := mock.Reader{
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
	file, err := f.ParseFile(path)
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
			mock.NewFileEntry("index.ts", false),
			mock.NewFileEntry("index.html", false),
			mock.NewFileEntry("components", true),
		},
		"/Users/conner/code/ore-ui/main": {
			mock.NewFileEntry("src", true),
			mock.NewFileEntry(".git", false),
		},
		"/Users/conner/code/ore-ui": {
			mock.NewFileEntry("main", true),
			mock.NewFileEntry("dev", true),
			mock.NewFileEntry(".bare", true),
			mock.NewFileEntry(".git", false),
		},
	}

	fileSystemMock := mock.Reader{
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
	file, err := f.ParseFile(path)
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
			mock.NewFileEntry("index.ts", false),
			mock.NewFileEntry("index.html", false),
			mock.NewFileEntry("components", true),
		},
		"/Users/conner/code/ore-ui/main": {
			mock.NewFileEntry("src", true),
			mock.NewFileEntry(".git", false),
		},
		"/Users/conner/code/ore-ui": {
			mock.NewFileEntry("main", true),
			mock.NewFileEntry("dev", true),
			mock.NewFileEntry(".bare", true),
			mock.NewFileEntry(".git", false),
		},
	}

	fileSystemMock := mock.Reader{
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
	file, err := f.ParseFile(path)
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
			mock.NewFileEntry("init.lua", false),
			mock.NewFileEntry("keybindings.lua", false),
			mock.NewFileEntry("autocommands.lua", false),
		},
		"/Users/conner/code/dotfiles/editors": {
			mock.NewFileEntry("nvim", true),
		},
		"/Users/conner/code/dotfiles": {
			mock.NewFileEntry("editors", true),
			mock.NewFileEntry("bootstrap.sh", false),
			mock.NewFileEntry("install.sh", false),
			mock.NewFileEntry(".git", true),
		},
	}

	fileSystemMock := mock.Reader{
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
	file, err := f.ParseFile(path)
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
