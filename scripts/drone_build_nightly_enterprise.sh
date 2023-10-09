#!/usr/bin/env bash
set -e
ver=$(cat ${GRAFANA_DIR}/package.json | jq -r .version)
local_dst="file://${DRONE_WORKSPACE}/dist"

# Check if version has hyphen
if [[ $ver == *-* ]]; then
    # If it does, replace everything after the hyphen
    ver=$(echo $ver | sed -E "s/-.*/-nightly.${DRONE_COMMIT_SHA:0:8}/")
else
    # If it doesn't, append "-nightly.${DRONE_COMMIT_SHA:0:8}"
    ver="${ver}-nightly.${DRONE_COMMIT_SHA:0:8}"
fi

# This command enables qemu emulators for building Docker images for arm64/armv6/armv7/etc on the host.
docker run --privileged --rm tonistiigi/binfmt --install all

# Build all of the grafana.tar.gz packages.
echo "Building tar.gz packages..."
dagger run --silent go run ./cmd \
  package \
  --grafana=false \
  --enterprise \
  --yarn-cache=${YARN_CACHE_FOLDER} \
  --distro=linux/amd64 \
  --distro=linux/arm64 \
  --distro=linux/arm/v6 \
  --distro=linux/arm/v7 \
  --distro=windows/amd64 \
  --distro=darwin/amd64 \
  --checksum \
  --build-id=${DRONE_BUILD_NUMBER} \
  --grafana-repo=https://github.com/grafana/grafana-security-mirror.git \
  --github-token=${GITHUB_TOKEN} \
  --go-version=${GO_VERSION} \
  --version=${ver} \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} >> assets.txt

# Use the non-windows, non-darwin, non-rpi packages and create deb packages from them.
dagger run --silent go run ./cmd deb \
  $(cat assets.txt | grep enterprise | grep tar.gz | grep -v docker | grep -v sha256 | grep -v windows | grep -v darwin | grep -v arm-6 | awk '{print "--package=" $0}') \
  --name="grafana-enterprise-nightly" \
  --checksum \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} >> assets.txt

# Use the armv7 package to build the `rpi` specific version.
dagger run --silent go run ./cmd deb \
  $(cat assets.txt | grep enterprise | grep tar.gz | grep -v docker | grep -v sha256 | grep -v windows | grep -v darwin | grep arm-7 | awk '{print "--package=" $0}') \
  --name="grafana-enterprise-nightly-rpi" \
  --checksum \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} >> assets.txt

# Make rpm installers for all the same Linux distros, and sign them because RPM packages are signed.
dagger run --silent go run ./cmd rpm \
  $(cat assets.txt | grep enterprise | grep tar.gz | grep -v docker | grep -v sha256 | grep -v windows | grep -v darwin | grep -v arm-6 | awk '{print "--package=" $0}') \
  --name="grafana-enterprise-nightly" \
  --checksum \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} \
  --sign=true \
  --gpg-private-key-base64="${GPG_PRIVATE_KEY}" \
  --gpg-public-key-base64="${GPG_PUBLIC_KEY}" \
  --gpg-passphrase="${GPG_PASSPHRASE}" >> assets.txt

# For Windows we distribute zips and exes
dagger run --silent go run ./cmd zip \
  $(cat assets.txt | grep enterprise | grep tar.gz | grep -v docker | grep -v sha256 | grep windows | awk '{print "--package=" $0}') \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} \
  --checksum >> assets.txt

dagger run --silent go run ./cmd windows-installer \
  $(cat assets.txt | grep enterprise | grep tar.gz | grep -v docker | grep -v sha256 | grep windows | awk '{print "--package=" $0}') \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} \
  --checksum >> assets.txt

# Build a docker image for all Linux distros except armv6
dagger run --silent go run ./cmd docker \
  $(cat assets.txt | grep enterprise | grep tar.gz | grep -v docker | grep -v sha256 | grep -v windows | grep -v darwin | grep -v arm-6 | awk '{print "--package=" $0}') \
  --checksum \
  --repo="grafana-enterprise-dev" \
  --ubuntu-base="ubuntu:22.04" \
  --alpine-base="alpine:3.18.0" \
  --destination=${local_dst} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} >> assets.txt

echo "Final list of artifacts:"
cat assets.txt
