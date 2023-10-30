#!/usr/bin/env bash
set -e
local_dst="${DRONE_WORKSPACE}/dist"

# This command enables qemu emulators for building Docker images for arm64/armv6/armv7/etc on the host.
docker run --privileged --rm tonistiigi/binfmt --install all

dagger run --silent go run ./cmd \
  package \
  -a targz:grafana:linux/amd64 \
  -a targz:grafana:linux/arm64 \
  -a targz:grafana:linux/arm/v7 \
  -a targz:grafana:linux/arm/v6 \
  -a deb:grafana:linux/amd64 \
  -a deb:grafana:linux/arm64 \
  -a deb:grafana:linux/arm/v6 \
  -a deb:grafana:linux/arm/v7 \
  -a rpm:grafana:linux/amd64 \
  -a rpm:grafana:linux/arm64 \
  -a targz:grafana:windows/amd64 \
  -a targz:grafana:windows/arm64 \
  -a targz:grafana:darwin/amd64 \
  -a targz:grafana:darwin/arm64 \
  -a zip:grafana:windows/amd64 \
  -a exe:grafana:windows/amd64 \
  -a docker:grafana:linux/amd64 \
  -a docker:grafana:linux/arm64 \
  -a docker:grafana:linux/arm/v7 \
  -a docker:grafana:linux/amd64:ubuntu \
  -a docker:grafana:linux/arm64:ubuntu \
  -a docker:grafana:linux/arm/v7:ubuntu \
  --checksum \
  --build-id=${DRONE_BUILD_NUMBER} \
  --grafana-dir=${GRAFANA_DIR} \
  --github-token=${GITHUB_TOKEN} \
  --go-version=${GO_VERSION} \
  --destination=${local_dst} \
  --yarn-cache=${YARN_CACHE_FOLDER} \
  --ubuntu-base="ubuntu:22.04" \
  --alpine-base="alpine:3.18.0" \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} >> assets.txt

cat assets.txt
