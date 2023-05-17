# Build & Packaging Grafana

The main goal of grafana-build is (as the name already indicates) building Grafana for various platforms. 
This actually consists of various parts as you need to have a binary of Grafana and the JavaScript/CSS frontend before you can then package everything up.

All of that is encompassed by the `package` command:

```
$ go run ./cmd package --distro darwin/arm64 --enterprise
```

The command above will build the backend binary for macOS on Apple Silicon and package that up into a single archive with the frontend artefacts: `grafana-enterprise-darwin-arm64.tar.gz`

If you then extract that package and run `./bin/grafana-server`, Grafana will launch and you will be able to access it via <http://localhost:3000>.

## Local checkout

If you want to use a local checkout of Grafana (for instance if you want to build it to test some change you made), then set the `--grafana-dir` flag accordingly.

The following command will create a binary package for `darwin/arm64` of Grafana based on a checkout inside the `$HOME/src/github.com/grafana/grafana` folder:

```
$ go run ./cmd package --distro darwin/arm64 --grafana-dir $HOME/src/github.com/grafana/grafana
```
