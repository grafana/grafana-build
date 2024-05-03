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

func AddFilesAndDirs(basePath string, client *dagger.Client, src *dagger.Directory, include git.Paths) *dagger.Directory {
	for _, path := range include.Files {
		filePath := basePath + path
		file := client.Host().File(filePath)
		slog.Info("adding file to dagger directory", "path", filePath)
		src = src.WithFile(path, file)
	}
	for _, path := range include.Directories {
		dirPath := basePath + path
		dir := client.Host().Directory(dirPath)
		slog.Info("adding directory to dagger directory", "path", dirPath)
		src = src.WithDirectory(path, dir)
	}
	return src
}
