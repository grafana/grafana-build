package containers

import (
	"fmt"

	"dagger.io/dagger"
)

type DockerBuildOpts struct {
	// Dockerfile is the path to the dockerfile with the '-f' command.
	// If it's not provided, then the docker command will default to 'Dockerfile' in `pwd`.
	Dockerfile string

	// Tags are provided as the '-t' argument, and can include the registry domain as well as the repository.
	// Docker build supports building the same image with multiple tags.
	// You might want to also include a 'latest' version of the tag.
	Tags []string
	// BuildArgs are provided to the docker command as '--build-arg'
	BuildArgs []string

	// UnixSocket should be created with the 'd.Host().UnixSocket(...)' function.
	// Most of the time, you will use 'd.Host().UnixSocket("/var/run/docker.sock")', but this unfortunately won't work on Windows machines.
	// TODO: Support an option to use the docker HTTP server.
	UnixSocket *dagger.Socket

	// Before allows the caller to add file and directory mounts, environment variables, etc before 'docker build' is called.
	// This is where you should add your context using relative paths; the docker context will be provided as '.'
	Before func(*dagger.Container) *dagger.Container
}

func DockerBuild(d *dagger.Client, opts *DockerBuildOpts) *dagger.Container {
	builder := d.Container().From("docker").
		WithUnixSocket("/var/run/docker.sock", opts.UnixSocket).
		WithWorkdir("/src")

	if opts.Before != nil {
		builder = opts.Before(builder)
	}

	args := []string{"docker", "buildx", "build", "."}
	for _, v := range opts.BuildArgs {
		args = append(args, fmt.Sprintf("--build-arg=%s", v))
	}
	for _, v := range opts.Tags {
		args = append(args, "-t", v)
	}

	return builder.WithExec(args)
}
