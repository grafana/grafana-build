# Docker image artifact

```
$ dagger run go run ./cmd artifacts -a docker:enterprise:linux/amd64 -a docker:enterprise:linux/amd64:ubuntu
# Produces dist/grafana-enterprise-10.1.0-pre_lUJuyyVXnECr_linux_amd64.docker.tar.gz (Alpine)
# Produces dist/grafana-enterprise-10.1.0-pre_lUJuyyVXnECr_linux_amd64.ubuntu.docker.tar.gz (Ubuntu)
```

You can then load these files into your Docker engine using the `docker load` command.
