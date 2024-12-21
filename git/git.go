package git

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"regexp"

	"github.com/viccon/pulse"
)

var (
	bareRepoExp    = regexp.MustCompile("gitdir: (?P<GitDir>.*)/worktrees")
	regularRepoExp = regexp.MustCompile(`url = .*(?:/|:)(?P<RepoName>[^/]*?)\.git`)
)

var (
	ErrEmptyPath         = errors.New("path is empty string")
	ErrPathNotAFile      = errors.New("the path is not a file")
	ErrReachedRoot       = errors.New("we reached the root without finding a .git file or folder")
	ErrParseRepoPath     = errors.New("failed to parse repository path")
	ErrParseBareRepoPath = errors.New("failed to parse bare repository path")
)

// Reader is an abstraction for the reader.
type Reader interface {
	Dir(string) string
	ReadDir(string) ([]fs.DirEntry, error)
	ReadFile(string) ([]byte, error)
	IsFile(string) bool
}

type FileParser struct {
	Reader Reader
}

// New creates a new FileParser.
func New() FileParser {
	return FileParser{filereader{}}
}

// extractSubExp extracts a named subgroup from a regexp match.
func extractSubExp(re *regexp.Regexp, matches []string, subexp string) string {
	exp := matches[re.SubexpIndex(subexp)]
	// We should never have a mismatch here.
	if exp == "" {
		panic(fmt.Sprintf("subexpression %s not found in %v. Did you modify the regexp?", subexp, matches))
	}
	return exp
}

// extractRepositoryName extracts the name of the repository from a .git file.
func (f FileParser) extractBareRepositoryPath(filepath string) (string, error) {
	fileContent, err := f.Reader.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	matches := bareRepoExp.FindStringSubmatch(string(fileContent))

	if len(matches) == 0 {
		return "", ErrParseBareRepoPath
	}

	return extractSubExp(bareRepoExp, matches, "GitDir"), nil
}

// findGitFolder calls itself recursively until it finds a .git
// configuration or reaches the root of the filesystem. It returns
// the name of the repository by parsing the url of the origin.
func (f FileParser) findGitFolder(dir string) (string, error) {
	// Stop the recursion if we have reached the root.
	if dir == "/" {
		return "", ErrReachedRoot
	}

	// Read the directory entries.
	entries, err := f.Reader.ReadDir(dir)
	if err != nil {
		return "", err
	}

	// Check if any of the entries is the .git file/folder.
	for _, e := range entries {
		if e.Name() == ".git" {
			// When I work on projects with long-lived branches I use worktrees. If that
			// is the case the .git file will point to the path of the bare directory.
			if !e.IsDir() {
				return f.extractBareRepositoryPath(path.Join(dir, ".git"))
			}
			return path.Join(dir, ".git"), nil
		}
	}

	// If we didn't find the .git file/folder we'll continue up the path.
	return f.findGitFolder(f.Reader.Dir(dir))
}

// extractRepositoryName extracts the name of the repository by
// looking at the url. This solves potential issues that could
// occur if you were to clone a repository under a different name.
func (f FileParser) extractRepositoryName(dirPath string) (string, error) {
	fileContent, err := f.Reader.ReadFile(path.Join(dirPath, "config"))
	if err != nil {
		return "", err
	}

	matches := regularRepoExp.FindStringSubmatch(string(fileContent))
	if len(matches) == 0 {
		return "", ErrParseRepoPath
	}

	return extractSubExp(regularRepoExp, matches, "RepoName"), nil
}

// ParseFile returns a ParseFile struct from an absolute path. It will return an
// error if the path is empty, if the path is not a file or if it can't find
// a parent .git file or folder before it reaches the root of the file tree.
func (f FileParser) ParseFile(absolutePath, filetype string) (pulse.GitFile, error) {
	if absolutePath == "" {
		return pulse.GitFile{}, ErrEmptyPath
	}

	// It could be a temporary buffer or directory.
	if !f.Reader.IsFile(absolutePath) {
		return pulse.GitFile{}, ErrPathNotAFile
	}

	// Check if the file is under source control.
	gitFolderPath, err := f.findGitFolder(f.Reader.Dir(absolutePath))
	if err != nil {
		return pulse.GitFile{}, err
	}

	repositoryName, err := f.extractRepositoryName(gitFolderPath)
	if err != nil {
		return pulse.GitFile{}, err
	}

	pathFromGitFolder := absolutePath[len(gitFolderPath)-len(".git"):]
	path := fmt.Sprintf("%s/%s", repositoryName, pathFromGitFolder)

	// Tries to get the filetype from either the file extension or name.
	filename := filepath.Base(absolutePath)
	gitFile := pulse.GitFile{
		Name:       filename,
		Filetype:   filetype,
		Repository: repositoryName,
		Path:       path,
	}

	return gitFile, nil
}

func ParseFile(absolutePath, filetype string) (pulse.GitFile, error) {
	return New().ParseFile(absolutePath, filetype)
}
