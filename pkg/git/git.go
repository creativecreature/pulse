package git

import (
	"errors"
	"path"
	"regexp"
	"strings"

	"code-harvest.conner.dev/pkg/filesystem"
)

var (
	ErrReachedRoot                     = errors.New("we reached the root without finding a .git file or folder")
	ErrParseRepoPath                   = errors.New("failed to parse repository path")
	ErrParseBareRepoPath               = errors.New("failed to parse bare repository path")
	ErrRepositoryDirectoryNameMismatch = errors.New("could not extract relative path in repo")
)

// Helper function to extract a subexpression from a regex.
func extractSubExp(re *regexp.Regexp, matches []string, subexp string) string {
	exp := matches[re.SubexpIndex(subexp)]
	// We should never have a missmatch here.
	if exp == "" {
		panic("Could not extract named subgroup. Did you modify the regexp?")
	}
	return exp
}

// Extracts the path to the bare repository from the .git file.
func extractBareRepositoryPath(fs filesystem.FileSystem, filepath string) (string, error) {
	fileContent, err := fs.ReadFile(filepath)
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
func extractRepositoryName(fs filesystem.FileSystem, dirPath string) (string, error) {
	fileContent, err := fs.ReadFile(path.Join(dirPath, "config"))
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
func findGitFolder(fs filesystem.FileSystem, dir string) (string, error) {
	// Stop the recursion if we have reached the root.
	if dir == "/" {
		return "", ErrReachedRoot
	}

	// Read the directory entries.
	entries, err := fs.ReadDir(dir)
	if err != nil {
		return "", err
	}

	// Check if any of the entries is the .git file/folder.
	for _, e := range entries {
		if e.Name() == ".git" {
			// When I work on projects with long lived branches I use worktrees. If
			// that is the case the .git file will point to the path of the bare dir.
			if !e.IsDir() {
				return extractBareRepositoryPath(fs, path.Join(dir, ".git"))
			}
			return path.Join(dir, ".git"), nil
		}
	}

	// If we didn't find the .git file/folder we'll continue up the path
	return findGitFolder(fs, fs.Dir(dir))
}

// Takes an absolute path of a file and tries to extract the name of the repository that it resides in
func GetRepositoryFromPath(fs filesystem.FileSystem, path string) (string, error) {
	rootPath, err := findGitFolder(fs, fs.Dir(path))

	if err != nil {
		return "", err
	}

	return extractRepositoryName(fs, rootPath)
}

// Extracts the path to the file starting from the repository
func GetRelativePathFromRepo(path, repo string) (string, error) {
	index := strings.Index(path, repo)
	if index == -1 {
		return "", ErrRepositoryDirectoryNameMismatch
	}
	return path[index:], nil
}
