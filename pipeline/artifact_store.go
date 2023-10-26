package pipeline

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"

	"dagger.io/dagger"
)

// The Storer stores the result of artifacts.
type ArtifactStore interface {
	StoreFile(ctx context.Context, a *Artifact, file *dagger.File) error
	File(ctx context.Context, a *Artifact) (*dagger.File, error)

	StoreDirectory(ctx context.Context, a *Artifact, dir *dagger.Directory) error
	Directory(ctx context.Context, a *Artifact) (*dagger.Directory, error)

	Export(ctx context.Context, a *Artifact, destination string) error
	Exists(ctx context.Context, a *Artifact) (bool, error)
}

type MapArtifactStore struct {
	data *sync.Map
}

func (m *MapArtifactStore) StoreFile(ctx context.Context, a *Artifact, file *dagger.File) error {
	f, err := a.Handler.Filename(ctx)
	if err != nil {
		return err
	}

	m.data.Store(f, file)
	return nil
}

func (m *MapArtifactStore) File(ctx context.Context, a *Artifact) (*dagger.File, error) {
	f, err := a.Handler.Filename(ctx)
	if err != nil {
		return nil, err
	}

	v, ok := m.data.Load(f)
	if !ok {
		return nil, errors.New("not found")
	}

	return v.(*dagger.File), nil
}

func (m *MapArtifactStore) StoreDirectory(ctx context.Context, a *Artifact, dir *dagger.Directory) error {
	f, err := a.Handler.Filename(ctx)
	if err != nil {
		return err
	}

	m.data.Store(f, dir)
	return nil
}

func (m *MapArtifactStore) Directory(ctx context.Context, a *Artifact) (*dagger.Directory, error) {
	f, err := a.Handler.Filename(ctx)
	if err != nil {
		return nil, err
	}

	v, ok := m.data.Load(f)
	if !ok {
		return nil, errors.New("not found")
	}

	return v.(*dagger.Directory), nil
}

func (m *MapArtifactStore) Export(ctx context.Context, a *Artifact, dst string) error {
	path, err := a.Handler.Filename(ctx)
	if err != nil {
		return err
	}

	path = filepath.Join(dst, path)
	switch a.Type {
	case ArtifactTypeFile:
		f, err := m.File(ctx, a)
		if err != nil {
			return err
		}

		if _, err := f.Export(ctx, path); err != nil {
			return err
		}

		return nil
	case ArtifactTypeDirectory:
		f, err := m.Directory(ctx, a)
		if err != nil {
			return err
		}

		if _, err := f.Export(ctx, path); err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("unrecognized artifact type: %d", a.Type)
}

func (m *MapArtifactStore) Exists(ctx context.Context, a *Artifact) (bool, error) {
	path, err := a.Handler.Filename(ctx)
	if err != nil {
		return false, err
	}

	_, ok := m.data.Load(path)
	return ok, nil
}

func NewArtifactStore(log *slog.Logger) ArtifactStore {
	return StoreWithLogging(&MapArtifactStore{
		data: &sync.Map{},
	}, log)
}
