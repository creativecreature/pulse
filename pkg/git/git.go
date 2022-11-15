package git

import (
	"errors"
	"io/fs"
	"path"
	"regexp"
)

var (
	ErrReachedRoot                     = errors.New("we reached the root without finding a .git file or folder")
	ErrParseRepoPath                   = errors.New("failed to parse repository path")
	ErrParseBareRepoPath               = errors.New("failed to parse bare repository path")
	ErrRepositoryDirectoryNameMismatch = errors.New("could not extract relative path in repo")
)

type Filesystem interface {
	Dir(string) string
	ReadDir(string) ([]fs.DirEntry, error)
	ReadFile(string) ([]byte, error)
}

type Git struct {
	fsys Filesystem
}

func New(fsys Filesystem) *Git {
	return &Git{
		fsys: fsys,
	}
}

// Helper function to extract a subexpression from a regex.
func extractSubExp(re *regexp.Regexp, matches []string, subexp string) string {
	exp := matches[re.SubexpIndex(subexp)]
	// We should never have a mismatch here.
	if exp == "" {
		panic("Could not extract named subgroup. Did you modify the regexp?")
	}
	return exp
}

// Extracts the path to the bare repository from the .git file.
func (g *Git) extractBareRepositoryPath(filepath string) (string, error) {
	fileContent, err := g.fsys.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile("gitdir: (?P<GitDir>.*)/worktrees")
	matches := re.FindStringSubmatch(string(fileContent))

	if len(matches) == 0 {
		return "", ErrParseBareRepoPath
	}

	return extractSubExp(re, matches, "GitDir"), nil
}

// Extracts the actual name of the repository by looking at the url.
func (g *Git) extractRepositoryName(dirPath string) (string, error) {
	fileContent, err := g.fsys.ReadFile(path.Join(dirPath, "config"))
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile("url = .*/(?P<RepoName>.*).git")
	matches := re.FindStringSubmatch(string(fileContent))

	if len(matches) == 0 {
		return "", ErrParseRepoPath
	}

	return extractSubExp(re, matches, "RepoName"), nil
}

// Calls itself recursively until it finds a .git file/folder or reaches the root.
// If it finds a .git file/folder it will try to extract the name of the repository.
func (g *Git) findGitFolder(dir string) (string, error) {
	// Stop the recursion if we have reached the root.
	if dir == "/" {
		return "", ErrReachedRoot
	}

	// Read the directory entries.
	entries, err := g.fsys.ReadDir(dir)
	if err != nil {
		return "", err
	}

	// Check if any of the entries is the .git file/folder.
	for _, e := range entries {
		if e.Name() == ".git" {
			// When I work on projects with long-lived branches I use worktrees. If
			// that is the case the .git file will point to the path of the bare dir.
			if !e.IsDir() {
				return g.extractBareRepositoryPath(path.Join(dir, ".git"))
			}
			return path.Join(dir, ".git"), nil
		}
	}

	// If we didn't find the .git file/folder we'll continue up the path
	return g.findGitFolder(g.fsys.Dir(dir))
}

// GetRepositoryFromPath takes an absolute path of a file and tries to extract the name of the repository that it resides in
func (g *Git) GetRepositoryNameFromPath(path string) (string, error) {
	rootPath, err := g.findGitFolder(g.fsys.Dir(path))
	if err != nil {
		return "", err
	}

	return g.extractRepositoryName(rootPath)
}
