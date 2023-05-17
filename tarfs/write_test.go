package tarfs_test

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/grafana/grafana-build/tarfs"
)

func TestWrite(t *testing.T) {
	tmp := t.TempDir()
	dir := os.DirFS("testdir")

	path := filepath.Join(tmp, "test.tar.gz")
	_, err := tarfs.WriteFile(path, dir)
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal("expected file to be openable, but enountered an error", err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)

	if err := fs.WalkDir(dir, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			t.Fatalf("did not expect un-walkable path (%s): %s", path, err)
		}

		if path == "." {
			return nil
		}

		if _, err := fs.Stat(dir, path); err != nil {
			t.Fatalf("did not expect error from fs.Stat (%s): %s", path, err)
		}

		// if info.IsDir() {
		// 	return nil
		// }

		h, err := tr.Next()
		if err != nil {
			t.Fatalf("did not expect error from getting next file header (%s): %s / %t", path, err, errors.Is(err, io.EOF))
		}

		if h.Name != path {
			t.Fatalf("Expected file '%s' in archive, but got '%s'", path, h.Name)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
