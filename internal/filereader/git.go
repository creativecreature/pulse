package filereader

import (
	"path"
	"regexp"
)

var (
	bareRepositoryRegexp = regexp.MustCompile("gitdir: (?P<GitDir>.*)/worktrees")
	repositoryRegexp     = regexp.MustCompile("url = .*/(?P<RepoName>.*).git")
)

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
func (f FileReader) extractBareRepositoryPath(filepath string) (string, error) {
	fileContent, err := f.fsys.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	matches := bareRepositoryRegexp.FindStringSubmatch(string(fileContent))

	if len(matches) == 0 {
		return "", ErrParseBareRepoPath
	}

	return extractSubExp(bareRepositoryRegexp, matches, "GitDir"), nil
}

// Extracts the actual name of the repository by looking at the url.
func (f FileReader) extractRepositoryName(dirPath string) (string, error) {
	fileContent, err := f.fsys.ReadFile(path.Join(dirPath, "config"))
	if err != nil {
		return "", err
	}

	matches := repositoryRegexp.FindStringSubmatch(string(fileContent))

	if len(matches) == 0 {
		return "", ErrParseRepoPath
	}

	return extractSubExp(repositoryRegexp, matches, "RepoName"), nil
}

// Calls itself recursively until it finds a .git file/folder or reaches the root.
// If it finds a .git file/folder it will try to extract the name of the repository.
func (f FileReader) findGitFolder(dir string) (string, error) {
	// Stop the recursion if we have reached the root.
	if dir == "/" {
		return "", ErrReachedRoot
	}

	// Read the directory entries.
	entries, err := f.fsys.ReadDir(dir)
	if err != nil {
		return "", err
	}

	// Check if any of the entries is the .git file/folder.
	for _, e := range entries {
		if e.Name() == ".git" {
			// When I work on projects with long-lived branches I use worktrees. If
			// that is the case the .git file will point to the path of the bare dir.
			if !e.IsDir() {
				return f.extractBareRepositoryPath(path.Join(dir, ".git"))
			}
			return path.Join(dir, ".git"), nil
		}
	}

	// If we didn't find the .git file/folder we'll continue up the path
	return f.findGitFolder(f.fsys.Dir(dir))
}

func (f FileReader) RepositoryName(path string) (string, error) {
	rootPath, err := f.findGitFolder(f.fsys.Dir(path))
	if err != nil {
		return "", err
	}

	return f.extractRepositoryName(rootPath)
}
