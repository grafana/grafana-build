package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

func main() {
	var (
		tarball = strings.ReplaceAll(os.Args[1], "file://", "")
	)

	ctx := context.Background()
	client, err := dagger.Connect(ctx)
	if err != nil {
		panic(err)
	}
	dir := filepath.Dir(tarball)
	file := client.Host().Directory(dir).File(filepath.Base(tarball))
	out := filepath.Join(dir, "npm-artifacts")
	artifacts := containers.ExtractedArchive(client, file).Directory("npm-artifacts")

	if _, err := artifacts.Export(ctx, out); err != nil {
		panic(err)
	}

	entries, err := artifacts.Entries(ctx)
	if err != nil {
		panic(err)
	}

	for _, v := range entries {
		f := "file://" + filepath.Join(out, v)
		fmt.Fprintln(os.Stdout, f)
	}
}
