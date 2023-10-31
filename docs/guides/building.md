# Build & Packaging Grafana

The main goal of grafana-build is (as the name already indicates) building Grafana for various platforms. 
This actually consists of various parts as you need to have a binary of Grafana and the JavaScript/CSS frontend before you can then package everything up.

All of that is encompassed by the `artifacts` command, which accepts a list of artifacts and attempts to create them:

```
$ dagger run go run ./cmd artifacts -a targz:grafana:linux/amd64
```

The command above will build the backend binary for Linux on an AMD64-compatible CPU and package that up into a [single archive][tarball] with the frontend artifacts: `grafana-enterprise-10.1.0-pre_lUJuyyVXnECr_linux_amd64.tar.gz`

If you then extract that package (`tar -xvf *.tar.gz`) and run `./bin/grafana-server`, Grafana will launch and you will be able to access it via <http://localhost:3000>.

## Local checkout

If you want to use a local checkout of Grafana (for instance if you want to build it to test some change you made), then set the `--grafana-dir` flag accordingly.

The following command will create a binary package for `darwin/arm64` of Grafana based on a checkout inside the `$HOME/src/github.com/grafana/grafana` folder:

```
$ dagger run go run ./cmd artifacts -a targz:grafana:linux/amd64 --grafana-dir=$HOME/src/github.com/grafana/grafana
```

## Platform packages

Now that you have a Grafana tarball with the main binaries and the frontend assets you can continue creating a package for your target distribution, or skip the tarball step and go straight to your package installer of choice.
grafana-build supports a handful of these specific [artifact types](../artifact-types/index.md) but for this tutorial let's build a [Debian package][deb]:

```
$ dagger run go run ./cmd artifacts -a deb:grafana:linux/amd64 --grafana-dir=$HOME/src/github.com/grafana/grafana
```

This will produce `grafana_10.1.0-pre_lUJuyyVXnECr_linux_amd64.deb` within the `dist` folder.

[tarball]: ../artifact-types/tarball.md
[deb]: ../artifact-types/deb.md
