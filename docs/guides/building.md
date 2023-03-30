# Build & Packaging Grafana

The main goal of grafana-build is (as the name already indicates) building Grafana for various platforms. 
This actually consists of various parts as you need to have a binary of Grafana and the JavaScript/CSS frontend before you can then package everything up.

All of that is encompassed by the `package` command:

```
$ go run ./cmd --enterprise package --distro darwin/amd64
```

The command above will build the backend binary for macOS on Apple Silicon and package that up into a single archive with the frontend artefacts: `grafana-enterprise-darwin-arm64.tar.gz`

If you then extract that package and run `./bin/grafana-server`, Grafana will launch and you will be able to access it via <http://localhost:3000>.