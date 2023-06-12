#!/usr/bin/env bash
dst="${DESTINATION}/${DRONE_BUILD_EVENT}"
local_dst="file://dist/${DRONE_BUILD_EVENT}"

# This command enables qemu emulators for building Docker images for arm64/armv6/armv7/etc on the host.
docker run --privileged --rm tonistiigi/binfmt --install all

# Build all of the grafana.tar.gz packages.
go run ./cmd \
  package \
  --yarn-cache=${YARN_CACHE_FOLDER} \
  --distro=linux/amd64 \
  --distro=linux/arm64 \
  --distro=linux/arm/v6 \
  --distro=linux/arm/v7 \
  --distro=windows/amd64 \
  --distro=darwin/amd64 \
  --checksum \
  --enterprise \
  --enterprise-ref="main" \
  --build-id=${DRONE_BUILD_NUMBER} \
  --grafana-dir=${GRAFANA_DIR} \
  --github-token=${GITHUB_TOKEN} \
  --version=${DRONE_TAG} \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} > assets.txt

# Build the grafana-pro tar.gz package.
go run ./cmd \
  package \
  --distro=linux/amd64 \
  --distro=linux/arm64 \
  --distro=linux/arm/v7 \
  --checksum \
  --env GO_BUILD_TAGS=pro \
  --env WIRE_TAGS=pro \
  --go-tags=pro \
  --edition=pro \
  --enterprise \
  --enterprise-ref="main" \
  --grafana=false \
  --build-id=${DRONE_BUILD_NUMBER} \
  --grafana-dir=${GRAFANA_DIR} \
  --github-token=${GITHUB_TOKEN} \
  --version=${DRONE_TAG} \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} >> assets.txt

Use the non-pro, non-windows, non-darwin packages and create deb packages from them.
go run ./cmd deb \
  $(cat assets.txt | grep tar.gz | grep -v docker | grep -v sha256 | grep -v windows | grep -v darwin | awk '{print "--package=" $0}') \
  --checksum \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} >> assets.txt & \
# Make rpm installers for all the same Linux distros, and sign them because RPM packages are signed.
go run ./cmd rpm \
  $(cat assets.txt | grep tar.gz | grep -v docker | grep -v sha256 | grep -v windows | grep -v darwin | awk '{print "--package=" $0}') \
  --checksum \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} >> assets.txt & \
# For Windows we distribute zips and exes
go run ./cmd zip \
  $(cat assets.txt | grep tar.gz | grep -v docker | grep -v sha256 | grep windows | awk '{print "--package=" $0}') \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} \
  --checksum >> assets.txt \
go run ./cmd windows-installer \
  $(cat assets.txt | grep tar.gz | grep -v docker | grep -v sha256 | grep windows | awk '{print "--package=" $0}') \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} \
  --checksum >> assets.txt \
fg

# Build a docker iamge for all Linux distros except armv6
go run ./cmd docker \
  $(cat assets.txt | grep tar.gz | grep -v docker | grep -v sha256 | grep -v windows | grep -v darwin | grep -v arm-6 | awk '{print "--package=" $0}') \
  --checksum \
  --ubuntu-base="ubuntu:22.10" \
  --alpine-base="alpine:3.18.0" \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} >> assets.txt

# Copy only the linux/amd64 edition frontends into a separate folder
go run ./cmd cdn \
  $(cat assets.txt | grep tar.gz | grep pro | grep -v docker | grep -v sha256 | grep linux_amd64 | awk '{print "--package=" $0}') \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} >> assets.txt

# Move the tar.gz packages to their expected locations
cat assets.txt | DESTINATION=gs://grafana-prerelease-dev go run ./scripts/move_packages.go ./dist/prerelease
