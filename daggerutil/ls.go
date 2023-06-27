package daggerutil

import (
	"context"
	"fmt"
	"io"
	"strings"

	"dagger.io/dagger"
)

// LS prints the directory contents for debugging
func LS(ctx context.Context, w io.Writer, dir *dagger.Directory) {
	entries, err := dir.Entries(ctx)
	if err != nil {
		fmt.Fprintf(w, "error listing directory contents: %s", err.Error())
		return
	}

	fmt.Fprintln(w, strings.Join(entries, "\n"))
}
