package containers

import (
	"context"
	"path"

	"dagger.io/dagger"
)

// BuildPlugin builds a single plugin built-in Grafana plugin located at 'path' in 'src'.
// basically it just calls 'yarn build' and returns the 'dist' folder that is generated.
// Since the plugins link to other grafana packages it's important that we have all of the grafana source code and not just the plugin source code unfortunately.
func BuildPlugin(d *dagger.Client, src *dagger.Directory, pluginPath string, yarnInstall *dagger.Directory, nodeVersion string) *dagger.Directory {
	return NodeContainer(d, NodeImage(nodeVersion)).
		WithMountedDirectory("/src", src).
		WithDirectory("/src/.yarn", yarnInstall.Directory("/.yarn")).
		WithDirectory("/src/node_modules", yarnInstall.Directory("/node_modules")).
		WithWorkdir(path.Join("/src", pluginPath)).
		WithExec([]string{"yarn", "install", "--immutable"}).
		WithExec([]string{"yarn", "build"}).
		Directory("dist")
}

type Plugin struct {
	Name      string
	Directory *dagger.Directory
}

// BuildPlugins builds all plugins in the directory located at 'pluginsPath' in 'src'. Each sub-directory is assumed to be a plugin.
func BuildPlugins(ctx context.Context, d *dagger.Client, src *dagger.Directory, pluginsPath, nodeVersion string) ([]Plugin, error) {
	dir := src.Directory(pluginsPath)
	entries, err := dir.Entries(ctx)
	if err != nil {
		return nil, err
	}

	plugins := make([]Plugin, len(entries))
	for i, v := range entries {
		modules := YarnInstall(d, src, nodeVersion)
		plugins[i] = Plugin{
			Name: v,
			// In a normal situation we would provide the directory as 'dir' and simply use the sub-path of 'v' but the plugins need the entire source tree of Grafana.
			Directory: BuildPlugin(d, src, path.Join(pluginsPath, v), modules, nodeVersion),
		}
	}

	return plugins, nil
}
