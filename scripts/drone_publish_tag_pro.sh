#!/usr/bin/env bash
dst="${DESTINATION}/${DRONE_BUILD_EVENT}"
local_dst="file://dist/${DRONE_BUILD_EVENT}"
set -e

# This command enables qemu emulators for building Docker images for arm64/armv6/armv7/etc on the host.
docker run --privileged --rm tonistiigi/binfmt --install all

# Build all of the grafana.tar.gz packages.
go run ./cmd \
  package \
  --yarn-cache=${YARN_CACHE_FOLDER} \
  --distro=linux/amd64 \
  --env GO_BUILD_TAGS=pro \
  --env WIRE_TAGS=pro \
  --go-tags=pro \
  --edition=pro \
  --checksum \
  --enterprise \
  --grafana=false \
  --build-id=${DRONE_BUILD_NUMBER} \
  --enterprise-dir=${GRAFANA_DIR} \
  --grafana-ref=${DRONE_TAG} \
  --grafana-repo=https://github.com/grafana/grafana-private-mirror.git \
  --github-token=${GITHUB_TOKEN} \
  --version=${DRONE_TAG} \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} > assets.txt

echo "Done building tar.gz packages..."

# Use the .tar.gz packages and create deb packages from them.
go run ./cmd deb \
  $(cat assets.txt | grep tar.gz | grep -v docker | grep -v sha256 | awk '{print "--package=" $0}') \
  --checksum \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} > debs.txt &

# Build a docker image for all .tar.gz packages
go run ./cmd docker \
  $(cat assets.txt | grep tar.gz | grep -v docker | grep -v sha256 | awk '{print "--package=" $0}') \
  --checksum \
  --ubuntu-base="ubuntu:22.10" \
  --alpine-base="alpine:3.18.0" \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} > docker.txt &

# Copy only the linux/amd64 edition frontends into a separate folder
go run ./cmd cdn \
  $(cat assets.txt | grep tar.gz | grep -v docker | grep -v sha256 | awk '{print "--package=" $0}') \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} > cdn.txt &

wait

cat debs.txt docker.txt cdn.txt >> assets.txt

# Move the tar.gz packages to their expected locations
cat assets.txt | DESTINATION=gs://grafana-prerelease-dev go run ./scripts/move_packages.go ./dist/prerelease
