#!/usr/bin/env bash
local_dst="dist/${DRONE_BUILD_EVENT}"
set -e

# This command enables qemu emulators for building Docker images for arm64/armv6/armv7/etc on the host.
docker run --privileged --rm tonistiigi/binfmt --install all

# Build all of the grafana.tar.gz packages.
dagger run --silent go run ./cmd \
  artifacts \
  -a frontend:enterprise \
  -a targz:pro:linux/amd64 \
  -a targz:pro:linux/arm64 \
  -a targz:pro:linux/arm/v6 \
  -a targz:pro:linux/arm/v7 \
  -a deb:pro:linux/amd64 \
  -a deb:pro:linux/arm64 \
  -a targz:pro:darwin/amd64 \
  -a targz:pro:windows/amd64 \
  --verify \
  --checksum \
  --parallel=2 \
  --yarn-cache=${YARN_CACHE_FOLDER} \
  --build-id=${DRONE_BUILD_NUMBER} \
  --enterprise-ref=${DRONE_TAG} \
  --grafana-ref=${DRONE_TAG} \
  --grafana-repo=https://github.com/grafana/grafana-security-mirror.git \
  --github-token=${GITHUB_TOKEN} \
  --version=${DRONE_TAG} \
  --go-version=${GO_VERSION} \
  --ubuntu-base="${UBUNTU_BASE}" \
  --alpine-base="${ALPINE_BASE}" \
  --destination=${local_dst} > assets.txt

touch "dist/grafana-pro_${DRONE_TAG}_${DRONE_BUILD_ID}_linux_amd64.ubuntu.docker.tar.gz"
touch "dist/grafana-pro_${DRONE_TAG}_${DRONE_BUILD_ID}_linux_arm64.ubuntu.docker.tar.gz"
touch "dist/grafana-pro_${DRONE_TAG}_${DRONE_BUILD_ID}_linux_arm-6.ubuntu.docker.tar.gz"
touch "dist/grafana-pro_${DRONE_TAG}_${DRONE_BUILD_ID}_linux_arm-7.ubuntu.docker.tar.gz"
touch "dist/grafana-pro_${DRONE_TAG}_${DRONE_BUILD_ID}_linux_amd64.docker.tar.gz"
touch "dist/grafana-pro_${DRONE_TAG}_${DRONE_BUILD_ID}_linux_arm64.docker.tar.gz"
touch "dist/grafana-pro_${DRONE_TAG}_${DRONE_BUILD_ID}_linux_arm-6.docker.tar.gz"
touch "dist/grafana-pro_${DRONE_TAG}_${DRONE_BUILD_ID}_linux_arm-7.docker.tar.gz"

echo "dist/grafana-pro_${DRONE_TAG}_${DRONE_BUILD_ID}_linux_amd64.ubuntu.docker.tar.gz" >> assets.txt
echo "dist/grafana-pro_${DRONE_TAG}_${DRONE_BUILD_ID}_linux_arm64.ubuntu.docker.tar.gz" >> assets.txt
echo "dist/grafana-pro_${DRONE_TAG}_${DRONE_BUILD_ID}_linux_arm-6.ubuntu.docker.tar.gz" >> assets.txt
echo "dist/grafana-pro_${DRONE_TAG}_${DRONE_BUILD_ID}_linux_arm-7.ubuntu.docker.tar.gz" >> assets.txt
echo "dist/grafana-pro_${DRONE_TAG}_${DRONE_BUILD_ID}_linux_amd64.docker.tar.gz" >> assets.txt
echo "dist/grafana-pro_${DRONE_TAG}_${DRONE_BUILD_ID}_linux_arm64.docker.tar.gz" >> assets.txt
echo "dist/grafana-pro_${DRONE_TAG}_${DRONE_BUILD_ID}_linux_arm-6.docker.tar.gz" >> assets.txt
echo "dist/grafana-pro_${DRONE_TAG}_${DRONE_BUILD_ID}_linux_arm-7.docker.tar.gz" >> assets.txt

# Move the tar.gz packages to their expected locations
cat assets.txt | go run ./scripts/move_packages.go ./dist/prerelease
