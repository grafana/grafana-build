package containers

import (
	"path"

	"dagger.io/dagger"
)

// BuildPlugin builds a single plugin built-in Grafana plugin located at 'path' in 'src'.
// basically it just calls 'yarn build' and returns the 'dist' folder that is generated.
// Since the plugins link to other grafana packages it's important that we have all of the grafana source code and not just the plugin source code unfortunately.
func BuildPlugin(d *dagger.Client, src *dagger.Directory, pluginPath, nodeVersion string) *dagger.Directory {
	return NodeContainer(d, NodeImage(nodeVersion)).
		WithMountedDirectory("/src", src).
		WithWorkdir(path.Join("/src", pluginPath)).
		WithExec([]string{"yarn", "install", "--immutable"}).
		WithExec([]string{"yarn", "build"}).
		Directory("dist")
}

type Plugin struct {
	Name      string
	Directory *dagger.Directory
}

/// // BuildAllPlugins builds all plugins in the directory located at 'src'. Each sub-directory is assumed to be a plugin.
/// func BuildAllPlugins(d *dagger.Client, src *dagger.Directory, nodeVersion string) []Plugin {
/// 	return nil
/// }
