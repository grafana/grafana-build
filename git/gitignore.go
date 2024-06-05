package git

import (
	"errors"
	"log/slog"
	"os"
	"strings"

	"dagger.io/dagger"
)

type Paths struct {
	Files       []string
	Directories []string
	Globs       []string
}

type Include struct {
	Files       map[string]dagger.File
	Directories map[string]dagger.Directory
}
type Gitignore struct {
	IncludePaths []string
	ExcludePaths []string
	Include      Paths
	Exclude      Paths
}

func IsDirectory(path string) (bool, bool, error) {
	fileInfo, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}

	return fileInfo.IsDir(), true, err
}

func DeterminePathType(path string, basePath string, paths *Paths) {
	// This is essentially the same as specifying the directory without the /*
	path = strings.TrimSuffix(path, "/*")
	if strings.Contains(path, "*") {
		slog.Info("found glob path", "path", path)
		paths.Globs = append(paths.Globs, path)
		return
	}
	cleanPath, _ := strings.CutPrefix(path, "/")
	cleanPath = basePath + cleanPath
	isDir, exists, err := IsDirectory(cleanPath)
	if err != nil {
		slog.Warn("failed to parse gitignore path", "path", cleanPath, "error", err)
		return
	}
	if exists {
		if isDir {
			slog.Info("found directory path", "path", path)
			paths.Directories = append(paths.Directories, path)
		} else {
			slog.Info("found file path", "path", path)
			paths.Files = append(paths.Files, path)
		}
	}
}

func ParseGitignore(basePath string, daggerDir *dagger.Directory, contents string) Gitignore {
	slog.Info("parsing gitignore contents")
	gitignore := Gitignore{}
	include := Paths{
		Files:       []string{},
		Directories: []string{},
		Globs:       []string{},
	}
	exclude := Paths{
		Files:       []string{},
		Directories: []string{},
		Globs:       []string{},
	}
	for _, line := range strings.Split(contents, "\n") {
		if line != "" {
			line = strings.TrimSpace(line)
			line, _ = strings.CutPrefix(line, "/")
			// Ignore comments
			if strings.HasPrefix(line, "#") {
				continue
			}
			if strings.HasPrefix(line, "!") {
				DeterminePathType(strings.TrimPrefix(line, "!"), basePath, &include)
			} else {
				DeterminePathType(line, basePath, &exclude)
			}
		}
	}

	allIncludePaths := []string{}
	allIncludePaths = append(allIncludePaths, include.Files...)
	allIncludePaths = append(allIncludePaths, include.Directories...)
	allIncludePaths = append(allIncludePaths, include.Globs...)
	allExcludePaths := []string{}
	allExcludePaths = append(allExcludePaths, exclude.Files...)
	allExcludePaths = append(allExcludePaths, exclude.Directories...)
	allExcludePaths = append(allExcludePaths, exclude.Globs...)

	gitignore.Include = include
	gitignore.IncludePaths = allIncludePaths
	gitignore.Exclude = exclude
	gitignore.ExcludePaths = allExcludePaths
	slog.Info("successfully parsed gitignore")
	return gitignore
}
