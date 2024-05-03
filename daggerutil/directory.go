package daggerutil

import (
	"log/slog"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/git"
)

func RemoveFilesAndDirs(src *dagger.Directory, paths git.Paths) *dagger.Directory {
	for _, path := range paths.Files {
		slog.Info("removing file from dagger directory", "path", path)
		src = src.WithoutFile(path)
	}
	for _, path := range paths.Directories {
		slog.Info("removing directory from dagger directory", "path", path)
		src = src.WithoutDirectory(path)
	}
	return src
}

func AddFilesAndDirs(src *dagger.Directory, include git.Include) *dagger.Directory {
	for path, file := range include.Files {
		file := file
		slog.Info("adding file to dagger directory", "path", path)
		src = src.WithFile(path, &file)
	}
	for path, dir := range include.Directories {
		dir := dir
		slog.Info("adding directory to dagger directory", "path", path)
		src = src.WithDirectory(path, &dir)
	}
	return src
}
