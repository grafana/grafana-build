#!/usr/bin/env sh
dst="${DESTINATION}/${DRONE_BUILD_EVENT}"
local_dst="file://dist/${DRONE_BUILD_EVENT}"
set -e

# This command enables qemu emulators for building Docker images for arm64/armv6/armv7/etc on the host.
docker run --privileged --rm tonistiigi/binfmt --install all

# Build all of the grafana.tar.gz packages.
dagger run --silent go run ./cmd \
  package \
  --yarn-cache=${YARN_CACHE_FOLDER} \
  --distro=linux/amd64 \
  --distro=linux/arm64 \
  --env GO_BUILD_TAGS=pro \
  --env WIRE_TAGS=pro \
  --go-tags=pro \
  --edition=pro \
  --checksum \
  --enterprise \
  --enterprise-ref=${ENTERPRISE_REF} \
  --grafana=false \
  --grafana-ref=${GRAFANA_REF} \
  --grafana-repo=https://github.com/grafana/grafana-security-mirror.git \
  --build-id=${DRONE_BUILD_NUMBER} \
  --github-token=${GITHUB_TOKEN} \
  --go-version=${GO_VERSION} \
  --version=$(cat package.json | jq -r .version | sed s/pre/${DRONE_BUILD_NUMBER}/g)
  --destination=${local_dst} > assets.txt

# Use the non-windows, non-darwin, non-rpi packages and create deb packages from them.
dagger run --silent go run ./cmd deb \
  $(cat assets.txt | grep tar.gz | grep -v docker | grep -v sha256 | grep -v windows | grep -v darwin | grep -v arm-6 | awk '{print "--package=" $0}') \
  --checksum \
  --destination=${local_dst} >> assets.txt

echo "Final list of artifacts:"
cat assets.txt
# Move the tar.gz packages to their expected locations
cat assets.txt | DESTINATION=gs://grafana-downloads IS_MAIN=true go run ./scripts/move_packages.go ./dist/main
