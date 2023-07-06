# Docker image artifact

```
$ go run ./cmd docker --package file://$PWD/dist/grafana-enterprise-10.1.0-pre_lUJuyyVXnECr_linux_amd64.tar.gz
# Produces dist/grafana-enterprise-10.1.0-pre_lUJuyyVXnECr_linux_amd64.ubuntu.docker.tar.gz (Ubuntu)
# Produces dist/grafana-enterprise-10.1.0-pre_lUJuyyVXnECr_linux_amd64.docker.tar.gz (Alpine)
```

You can then load these files into your Docker engine using the `docker load` command.
