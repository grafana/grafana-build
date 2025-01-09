#!/usr/bin/env sh
local_dst="dist/${DRONE_BUILD_EVENT}"
set -e

# This command enables qemu emulators for building Docker images for arm64/armv6/armv7/etc on the host.
docker run --privileged --rm tonistiigi/binfmt --install all
dagger run --silent go run ./cmd \
  artifacts \
  -a targz:enterprise:linux/amd64 \
  -a targz:enterprise:linux/arm64 \
  -a targz:enterprise:linux/arm/v6 \
  -a targz:enterprise:linux/arm/v7 \
  -a deb:enterprise:linux/amd64 \
  -a deb:enterprise:linux/arm64 \
  -a deb:enterprise:linux/arm/v6 \
  -a deb:enterprise:linux/arm/v7 \
  -a docker:enterprise:linux/amd64 \
  -a docker:enterprise:linux/arm64 \
  --yarn-cache=${YARN_CACHE_FOLDER} \
  --checksum \
  --verify \
  --build-id=${DRONE_BUILD_NUMBER} \
  --grafana-ref=${SOURCE_COMMIT} \
  --grafana-repo="https://github.com/grafana/grafana.git" \
  --enterprise-ref=${DRONE_COMMIT} \
  --github-token=${GITHUB_TOKEN} \
  --go-version=${GO_VERSION} \
  --ubuntu-base=${UBUNTU_BASE} \
  --alpine-base=${ALPINE_BASE} \
  --patches-repo=${PATCHES_REPO} \
  --patches-path=${PATCHES_PATH} \
  --destination=${local_dst} > assets.txt

cat assets.txt

# Move the tar.gz packages to their expected locations
cat assets.txt | DESTINATION=gs://grafana-downloads IS_MAIN=true go run ./scripts/move_packages.go ./dist/main
