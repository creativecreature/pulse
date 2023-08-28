package filereader

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"regexp"

	codeharvest "github.com/creativecreature/code-harvest"
	"github.com/creativecreature/code-harvest/filetypes"
)

var (
	bareRepoExp    = regexp.MustCompile("gitdir: (?P<GitDir>.*)/worktrees")
	regularRepoExp = regexp.MustCompile("url = .*/(?P<RepoName>.*).git")
)

var (
	ErrEmptyPath         = errors.New("path is empty string")
	ErrPathNotAFile      = errors.New("the path is not a file")
	ErrReachedRoot       = errors.New("we reached the root without finding a .git file or folder")
	ErrParseRepoPath     = errors.New("failed to parse repository path")
	ErrParseBareRepoPath = errors.New("failed to parse bare repository path")
)

type FileReader struct {
	Reader Reader
}

// New creates a new FileReader.
func New() FileReader {
	return FileReader{reader{}}
}

// extractSubExp extracts a named subgroup from a regexp match.
func extractSubExp(re *regexp.Regexp, matches []string, subexp string) string {
	exp := matches[re.SubexpIndex(subexp)]
	// We should never have a mismatch here.
	if exp == "" {
		panic("Could not extract named subgroup. Did you modify the regexp?")
	}
	return exp
}

// extractRepositoryName extracts the name of the repository from a .git file.
func (g FileReader) extractBareRepositoryPath(filepath string) (string, error) {
	fileContent, err := g.Reader.ReadFile(filepath)
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
func (g FileReader) findGitFolder(dir string) (string, error) {
	// Stop the recursion if we have reached the root.
	if dir == "/" {
		return "", ErrReachedRoot
	}

	// Read the directory entries.
	entries, err := g.Reader.ReadDir(dir)
	if err != nil {
		return "", err
	}

	// Check if any of the entries is the .git file/folder.
	for _, e := range entries {
		if e.Name() == ".git" {
			// When I work on projects with long-lived branches I use worktrees. If that
			// is the case the .git file will point to the path of the bare directory.
			if !e.IsDir() {
				return g.extractBareRepositoryPath(path.Join(dir, ".git"))
			}
			return path.Join(dir, ".git"), nil
		}
	}

	// If we didn't find the .git file/folder we'll continue up the path.
	return g.findGitFolder(g.Reader.Dir(dir))
}

// extractRepositoryName extracts the name of the repository by
// looking at the url. This solves potential issues that could
// occur if you were to clone a repository under a different name.
func (git FileReader) extractRepositoryName(dirPath string) (string, error) {
	fileContent, err := git.Reader.ReadFile(path.Join(dirPath, "config"))
	if err != nil {
		return "", err
	}

	matches := regularRepoExp.FindStringSubmatch(string(fileContent))

	if len(matches) == 0 {
		return "", ErrParseRepoPath
	}

	return extractSubExp(regularRepoExp, matches, "RepoName"), nil
}

// GitFile returns a GitFile struct from an absolute path. It will return an
// error if the path is empty, if the path is not a file or if it can't find
// a parent .git file or folder before it reaches the root of the file tree.
func (g FileReader) GitFile(absolutePath string) (codeharvest.GitFile, error) {
	if absolutePath == "" {
		return codeharvest.GitFile{}, ErrEmptyPath
	}

	// It could be a temporary buffer or directory.
	if !g.Reader.IsFile(absolutePath) {
		return codeharvest.GitFile{}, ErrPathNotAFile
	}

	// Check if the file is under source control.
	gitFolderPath, err := g.findGitFolder(g.Reader.Dir(absolutePath))
	if err != nil {
		return codeharvest.GitFile{}, err
	}

	repositoryName, err := g.extractRepositoryName(gitFolderPath)
	if err != nil {
		return codeharvest.GitFile{}, err
	}

	pathFromGitFolder := absolutePath[len(gitFolderPath)-len(".git"):]
	path := fmt.Sprintf("%s/%s", repositoryName, pathFromGitFolder)

	// Tries to get the filetype from either the file extension or name.
	filename := filepath.Base(absolutePath)
	ft, err := filetypes.Type(filename)
	if err != nil {
		return codeharvest.GitFile{}, err
	}

	gitFile := codeharvest.GitFile{
		Name:       filename,
		Filetype:   ft,
		Repository: repositoryName,
		Path:       path,
	}

	return gitFile, nil
}
