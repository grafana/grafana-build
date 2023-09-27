#!/usr/bin/env sh
local_dst="file://dist/${DRONE_BUILD_EVENT}"
set -e

# This command enables qemu emulators for building Docker images for arm64/armv6/armv7/etc on the host.
docker run --privileged --rm tonistiigi/binfmt --install all

dagger run --silent go run ./cmd \
  package \
  --distro=linux/amd64 \
  --distro=linux/arm64 \
  --version=$(echo ${GRAFANA_VERSION} | sed s/pre/${DRONE_BUILD_NUMBER}/g) \
  --grafana=false \
  --grafana-ref=${GRAFANA_REF} \
  --grafana-repo=https://github.com/grafana/grafana-security-mirror.git \
  --enterprise \
  --enterprise-ref=${ENTERPRISE_REF} \
  --checksum \
  --build-id=${DRONE_BUILD_NUMBER} \
  --github-token=${GITHUB_TOKEN} \
  --go-version=${GO_VERSION} \
  --destination=${local_dst} > assets.txt \

# Use the non-windows, non-darwin, non-rpi packages and create deb packages from them.
dagger run --silent go run ./cmd deb \
  $(cat assets.txt | grep tar.gz | grep -v docker | grep -v sha256 | grep -v windows | grep -v darwin | grep -v arm-6 | awk '{print "--package=" $0}') \
  --checksum \
  --destination=${local_dst} >> assets.txt

echo "Final list of artifacts:"
cat assets.txt

# Move the tar.gz packages to their expected locations
cat assets.txt | DESTINATION=gs://grafana-downloads IS_MAIN=true go run ./scripts/move_packages.go ./dist/main
