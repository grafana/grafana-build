# Windows installer artifact (.exe)

grafana-build can create a Windows installer out of a [Grafana tarball][pkg] using [NSIS][].

```
$ go run ./cmd windows-installer --package file://$PWD/dist/grafana.tar.gz
```

[nsis]: https://nsis.sourceforge.io/Main_Page
[pkg]: ./tarball.md
