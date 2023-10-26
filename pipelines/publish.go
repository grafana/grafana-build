package pipelines

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
	"golang.org/x/sync/semaphore"
)

type SyncWriter struct {
	Writer io.Writer

	mutex *sync.Mutex
}

func NewSyncWriter(w io.Writer) *SyncWriter {
	return &SyncWriter{
		Writer: w,
		mutex:  &sync.Mutex{},
	}
}

func (w *SyncWriter) Write(b []byte) (int, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.Writer.Write(b)
}

var Stdout = NewSyncWriter(os.Stdout)

func PublishFileFunc(ctx context.Context, sm *semaphore.Weighted, d *dagger.Client, opts *containers.PublishFileOpts) func() error {
	return func() error {
		log.Printf("[%s] Attempting to publish file", opts.Destination)
		log.Printf("[%s] Acquiring semaphore", opts.Destination)
		if err := sm.Acquire(ctx, 1); err != nil {
			return fmt.Errorf("failed to acquire semaphore: %w", err)
		}
		defer sm.Release(1)
		log.Printf("[%s] Acquired semaphore", opts.Destination)

		log.Printf("[%s] Publishing file", opts.Destination)
		out, err := containers.PublishFile(ctx, d, opts)
		if err != nil {
			return fmt.Errorf("[%s] error: %w", opts.Destination, err)
		}
		log.Printf("[%s] Done publishing file", opts.Destination)

		fmt.Fprintln(Stdout, strings.Join(out, "\n"))
		return nil
	}
}

func PublishDirFunc(ctx context.Context, sm *semaphore.Weighted, d *dagger.Client, dir *dagger.Directory, opts *containers.GCPOpts, dst string) func() error {
	return func() error {
		log.Printf("[%s] Attempting to publish file", dst)
		log.Printf("[%s] Acquiring semaphore", dst)
		if err := sm.Acquire(ctx, 1); err != nil {
			return fmt.Errorf("failed to acquire semaphore: %w", err)
		}
		defer sm.Release(1)
		log.Printf("[%s] Acquired semaphore", dst)

		log.Printf("[%s] Publishing file", dst)
		out, err := containers.PublishDirectory(ctx, d, dir, opts, dst)
		if err != nil {
			return fmt.Errorf("[%s] error: %w", dst, err)
		}
		log.Printf("[%s] Done publishing file", dst)

		fmt.Fprintln(Stdout, out)
		return nil
	}
}
