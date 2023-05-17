package tarfs

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
)

// WriteFile writes the filesystem (dir) into a gzipped tar archive at the name provided.
// This function closes the File.
func WriteFile(name string, dir fs.FS) (*os.File, error) {
	file, err := os.Create(name)
	if err != nil {
		return nil, fmt.Errorf("error creating tar.gz: %w", err)
	}
	defer file.Close()
	if err := Write(file, dir); err != nil {
		return nil, fmt.Errorf("error writing to tar.gz: %w", err)
	}

	return file, nil
}

// Write writes the filesystem (dir) into a gzipped tar archive in the writer provided.
func Write(writer io.Writer, dir fs.FS) error {
	gzw := gzip.NewWriter(writer)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	return fs.WalkDir(dir, ".", func(path string, d fs.DirEntry, err error) error {
		// If there's something that we can't package, like maybe a symbolic link, we should probably return an error.
		// In the future, should we try allow the user to define what to do?
		if err != nil {
			return fmt.Errorf("fs.WalkDir error argument: %w", err)
		}
		if path == "." {
			return nil
		}

		info, err := fs.Stat(dir, path)
		if err != nil {
			return err
		}

		h, err := tar.FileInfoHeader(info, path)
		if err != nil {
			return err
		}

		h.Name = path

		if err := tw.WriteHeader(h); err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := dir.Open(path)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, file); err != nil {
				return err
			}
			if err := file.Close(); err != nil {
				return err
			}
		}

		return nil
	})
}
