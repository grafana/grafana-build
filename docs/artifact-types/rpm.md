# RPM artifact (.rpm)

```
$ dagger run go run ./cmd artifacts -a rpm:enterprise:linux/amd64
# Produces dist/grafana-enterprise-10.1.0-pre_lUJuyyVXnECr_linux_amd64.rpm (Unisnged)

$ dagger run go run ./cmd artifacts -a rpm:enterprise:linux/amd64:sign
# Produces dist/grafana-enterprise-10.1.0-pre_lUJuyyVXnECr_linux_amd64.rpm (Unisnged)
```

If `GPG_PRIVATE_KEY`, `GPG_PUBLIC_KEY`, `GPG_PASSPHRASE` environment variables are set (and are base64 encoded), then the RPM will be signed if the `:sign` flag is added.

Example:

```
export GPG_PRIVATE_KEY=$(cat ./key.private | base64 -w 0)
export GPG_PUBLIC_KEY=$(cat ./key.public | base64 -w 0)
export GPG_PASSPHRASE=grafana


dagger run go run ./cmd artifacts -a rpm:enterprise:linux/amd64:sign
# Produces dist/grafana-enterprise-10.1.0-pre_lUJuyyVXnECr_linux_amd64.rpm (Signed)
```
