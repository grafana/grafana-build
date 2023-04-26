package containers

import "dagger.io/dagger"

const RubyContainer = "ruby:3.2.2-bullseye"

func FPMContainer(d *dagger.Client) *dagger.Container {
	return d.Container().
		From(RubyContainer).
		WithEntrypoint(nil).
		WithExec([]string{"gem", "install", "fpm"})
}
