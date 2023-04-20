package pipelines

import (
	"context"
	"log"

	"dagger.io/dagger"
	"github.com/grafana/grafana-build/containers"
)

var IntegrationDatabases = []string{"sqlite", "mysql", "postgres"}

// GrafanaBackendTests runs the Grafana backend test containers for short (unit) and integration tests.
func GrafanaBackendTests(ctx context.Context, d *dagger.Client, src *dagger.Directory, args PipelineArgs) error {
	db := args.Context.String("database")

	var r = []*dagger.Container{}
	c := d.Pipeline("backend tests", dagger.PipelineOpts{
		Description: "Runs backend unit tests",
	})
	if args.Context.Bool("unit") {
		// add unit tests to execution list
		log.Println("Unit tests will be run")
		r = append(r, containers.BackendTestShort(c, src))
	}
	if args.Context.Bool("integration") {
		// add integration tests to execution list
		log.Printf("Integration tests will be run using a '%s' database", db)
		r = append(r, containers.BackendTestIntegration(c, src))
	}
	return containers.Run(ctx, r)
}

// GrafanaBackendTestIntegration runs the Grafana backend test containers for short (unit) and integration tests.
func GrafanaBackendTestIntegration(ctx context.Context, d *dagger.Client, src *dagger.Directory, args PipelineArgs) error {
	c := d.Pipeline("integration tests", dagger.PipelineOpts{
		Description: "Runs backend integration tests",
	})

	r := []*dagger.Container{
		containers.BackendTestIntegration(c, src),
	}

	return containers.Run(ctx, r)
}
