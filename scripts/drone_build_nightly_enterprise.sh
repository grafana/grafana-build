#!/usr/bin/env bash
set -e
local_dst="${DRONE_WORKSPACE}/dist"

# This command enables qemu emulators for building Docker images for arm64/armv6/armv7/etc on the host.
docker run --privileged --rm tonistiigi/binfmt --install all
dagger run --silent go run ./cmd \
  artifacts \
  -a targz:enterprise:linux/amd64 \
  -a targz:enterprise:linux/arm64 \
  # -a targz:enterprise:linux/arm/v6 \
  # -a targz:enterprise:linux/arm/v7 \
  -a deb:enterprise:linux/amd64 \
  -a deb:enterprise:linux/arm64 \
  # -a deb:enterprise:linux/arm/v6 \
  # -a deb:enterprise:linux/arm/v7 \
  -a rpm:enterprise:linux/amd64 \
  -a rpm:enterprise:linux/arm64 \
  -a targz:enterprise:windows/amd64 \
  -a targz:enterprise:windows/arm64 \
  -a zip:enterprise:windows/amd64 \
  -a zip:enterprise:windows/arm64 \
  -a targz:enterprise:darwin/amd64 \
  -a targz:enterprise:darwin/arm64 \
  -a exe:enterprise:windows/amd64 \
  -a docker:enterprise:linux/amd64 \
  -a docker:enterprise:linux/arm64 \
  -a docker:enterprise:linux/amd64:ubuntu \
  -a docker:enterprise:linux/arm64:ubuntu \
  --yarn-cache=${YARN_CACHE_FOLDER} \
  --checksum \
  --verify \
  --build-id=${DRONE_BUILD_NUMBER} \
  --grafana-ref=main \
  --enterprise-ref=main \
  --grafana-repo=https://github.com/grafana/grafana-security-mirror.git \
  --github-token=${GITHUB_TOKEN} \
  --go-version=${GO_VERSION} \
  --destination=${local_dst} \
  --ubuntu-base="${UBUNTU_BASE}" \
  --alpine-base="${ALPINE_BASE}" > assets.txt

cat assets.txt
