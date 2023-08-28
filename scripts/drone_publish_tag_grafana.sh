#!/usr/bin/env bash
dst="${DESTINATION}/${DRONE_BUILD_EVENT}"
local_dst="file://dist/${DRONE_BUILD_EVENT}"
set -e

# This command enables qemu emulators for building Docker images for arm64/armv6/armv7/etc on the host.
docker run --privileged --rm tonistiigi/binfmt --install all

# Build all of the grafana.tar.gz packages.
echo "Building tar.gz packages..."
dagger run --silent go run ./cmd \
  package \
  --yarn-cache=${YARN_CACHE_FOLDER} \
  --distro=linux/amd64 \
  --distro=linux/arm64 \
  --distro=linux/arm/v6 \
  --distro=linux/arm/v7 \
  --distro=windows/amd64 \
  --distro=darwin/amd64 \
  --checksum \
  --build-id=${DRONE_BUILD_NUMBER} \
  --grafana-dir=${GRAFANA_DIR} \
  --github-token=${GITHUB_TOKEN} \
  --version=${DRONE_TAG} \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} > assets.txt

echo "Done building tar.gz packages..."
cat assets.txt

# Create the npm artifacts using only the amd64 linux package
dagger run --silent go run ./scripts/copy_npm $(cat assets.txt | grep tar.gz | grep linux | grep amd64 | grep -v sha256 -m 1) > npm.txt

# Copy only the linux/amd64 edition storybook into a separate folder
dagger run --silent go run ./cmd storybook \
  $(cat assets.txt | grep tar.gz | grep linux | grep amd64 | grep -v sha256 | awk '{print "--package=" $0}') \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} > storybook.txt

# Use the non-windows, non-darwin, non-rpi packages and create deb packages from them.
dagger run --silent go run ./cmd deb \
  $(cat assets.txt | grep tar.gz | grep -v docker | grep -v sha256 | grep -v windows | grep -v darwin | grep -v arm-7 | grep -v arm-6 | awk '{print "--package=" $0}') \
  --parallel=16 \
  --checksum \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} > debs.txt

# Use the armv7 package to build the `rpi` specific version.
dagger run --silent go run ./cmd deb \
  $(cat assets.txt | grep tar.gz | grep -v docker | grep -v sha256 | grep -v windows | grep -v darwin | grep arm-7 | awk '{print "--package=" $0}') \
  --parallel=16 \
  --name=grafana-rpi \
  --checksum \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} >> debs.txt

# Make rpm installers for all the same Linux distros, and sign them because RPM packages are signed.
dagger run --silent go run ./cmd rpm \
  $(cat assets.txt | grep tar.gz | grep -v docker | grep -v sha256 | grep -v windows | grep -v darwin | grep -v arm-6 | awk '{print "--package=" $0}') \
  --checksum \
  --parallel=16 \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} \
  --sign=true \
  --gpg-private-key-base64="${GPG_PRIVATE_KEY}" \
  --gpg-public-key-base64="${GPG_PUBLIC_KEY}" \
  --gpg-passphrase="${GPG_PASSPHRASE}" > rpms.txt

# For Windows we distribute zips and exes
dagger run --silent go run ./cmd zip \
  $(cat assets.txt | grep tar.gz | grep -v docker | grep -v sha256 | grep windows | awk '{print "--package=" $0}') \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} \
  --checksum > zips.txt

dagger run --silent go run ./cmd windows-installer \
  $(cat assets.txt | grep tar.gz | grep -v docker | grep -v sha256 | grep windows | awk '{print "--package=" $0}') \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} \
  --checksum > exes.txt

# Build a docker image for all Linux distros except armv6
dagger run --silent go run ./cmd docker \
  $(cat assets.txt | grep tar.gz | grep -v docker | grep -v sha256 | grep -v windows | grep -v darwin | grep -v arm-6 | awk '{print "--package=" $0}') \
  --checksum \
  --ubuntu-base="ubuntu:22.10" \
  --alpine-base="alpine:3.18.0" \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} > docker.txt

cat debs.txt rpms.txt zips.txt exes.txt docker.txt npm.txt storybook.txt >> assets.txt

echo "Final list of artifacts:"
cat assets.txt

# Move the tar.gz packages to their expected locations
cat assets.txt | go run ./scripts/move_packages.go ./dist/prerelease
