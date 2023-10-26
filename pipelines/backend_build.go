package pipelines

import (
	"context"

	"dagger.io/dagger"
)

// GrafanaBackendBuild builds all of the distributions in the '--distros' argument and places them in the 'bin' directory of the PWD.
func GrafanaBackendBuild(ctx context.Context, d *dagger.Client, src *dagger.Directory, args PipelineArgs) error {
	// var (
	// 	distroList = args.Context.StringSlice("distro")
	// 	distros    = make([]executil.Distribution, len(distroList))
	// )

	// for i, v := range distroList {
	// 	distros[i] = executil.Distribution(v)
	// }

	// dirs := make([]*dagger.Directory, len(distroList))
	// for i, distro := range distros {
	// 	opts := &backend.GrafanaCompileOpts{
	// 		Source:           src,
	// 		Distribution:     distro,
	// 		Platform:         args.Platform,
	// 		Version:          args.GrafanaOpts.Version,
	// 		Env:              args.GrafanaOpts.Env,
	// 		GoTags:           args.GrafanaOpts.GoTags,
	// 		GoVersion:        args.GrafanaOpts.GoVersion,
	// 		YarnCacheHostDir: args.GrafanaOpts.YarnCacheHostDir,
	// 	}
	// 	builder := containers.GolangContainer(d, opts.Platform, fmt.Sprintf("golang:%s-alpine", args.GrafanaOpts.GoVersion))
	// 	dir := containers.BackendBinDir(builder, distro)
	// 	dirs[i] = dir
	// }

	// for i, v := range distros {
	// 	var (
	// 		dir    = dirs[i]
	// 		output = filepath.Join("bin", string(v))
	// 	)
	// 	if _, err := dir.Export(ctx, output); err != nil {
	// 		return err
	// 	}
	// }
	// return nil
	return nil
}
