package filereader

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"regexp"

	"code-harvest.conner.dev/filetypes"
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

func New() FileReader {
	return FileReader{reader{}}
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

// Calls itself recursively until it finds a .git file/folder or reaches the root.
// If it finds a .git file/folder it will try to extract the name of the repository.
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
			// When I work on projects with long-lived branches I use worktrees. If
			// that is the case the .git file will point to the path of the bare dir.
			if !e.IsDir() {
				return g.extractBareRepositoryPath(path.Join(dir, ".git"))
			}
			return path.Join(dir, ".git"), nil
		}
	}

	// If we didn't find the .git file/folder we'll continue up the path
	return g.findGitFolder(g.Reader.Dir(dir))
}

// Extracts the actual name of the repository by looking at the url.
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

func (g FileReader) GitFile(absolutePath string) (File, error) {
	if absolutePath == "" {
		return file{}, ErrEmptyPath
	}

	// It could be a temporary buffer or directory.
	if !g.Reader.IsFile(absolutePath) {
		return file{}, ErrPathNotAFile
	}

	// When I aggregate the data I do it on a per project basis. Therefore, if this
	// is just a one-off edit of some configuration file I won't track time for it.
	gitFolderPath, err := g.findGitFolder(g.Reader.Dir(absolutePath))
	if err != nil {
		return file{}, err
	}

	repositoryName, err := g.extractRepositoryName(gitFolderPath)
	if err != nil {
		return file{}, err
	}

	pathFromGitFolder := absolutePath[len(gitFolderPath)-len(".git"):]
	path := fmt.Sprintf("%s/%s", repositoryName, pathFromGitFolder)

	// Tries to get the filetype from either the file extension or name.
	filename := filepath.Base(absolutePath)
	ft, err := filetypes.Type(filename)
	if err != nil {
		return file{}, err
	}

	fileMetaData := file{filename, ft, repositoryName, path}
	return fileMetaData, nil
}
