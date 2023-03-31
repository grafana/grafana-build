package pipelines

import (
	"context"
	"dagger.io/dagger"
	"fmt"
	"github.com/grafana/grafana-build/containers"
	"log"
	"strings"
)

var IntegrationDatabases = []string{"sqlite", "mysql", "postgres"}

// list of strings, string
func Contains(arr []string, v string) bool {
	for _, str := range arr {
		if str == v {
			return true
		}
	}
	return false
}

// GrafanabackendTests runs the Grafana backend test containers for short (unit) and integration tests.
func GrafanaBackendTests(ctx context.Context, d *dagger.Client, args PipelineArgs) error {

	db := args.Context.String("database")
	log.Println("db", db)

	// verify the database is one of the available choices
	if !Contains(IntegrationDatabases, db) {
		return fmt.Errorf("database needs to be one of %s", strings.Join(IntegrationDatabases, ","))
	}
	var r = []*dagger.Container{}
	c := d.Pipeline("backend tests", dagger.PipelineOpts{
		Description: "Runs backend unit tests",
	})
	if args.Context.Bool("unit") == true {
		// add unit tests to execution list
		log.Println("Unit tests will be run")
		r = append(r, containers.BackendTestShort(c, args.Grafana))
	}
	if args.Context.Bool("integration") == true {
		// add integration tests to execution list
		log.Printf("Integration tests will be run using a '%s' database", db)
		r = append(r, containers.BackendTestIntegration(c, args.Grafana))
	}
	return containers.Run(ctx, r)
}

// GrafanabackendTestIntegration runs the Grafana backend test containers for short (unit) and integration tests.
func GrafanaBackendTestIntegration(ctx context.Context, d *dagger.Client, args PipelineArgs) error {
	c := d.Pipeline("integration tests", dagger.PipelineOpts{
		Description: "Runs backend integration tests",
	})

	r := []*dagger.Container{
		containers.BackendTestIntegration(c, args.Grafana),
	}

	return containers.Run(ctx, r)
}
