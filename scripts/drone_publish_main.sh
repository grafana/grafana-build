#!/usr/bin/env sh

# This command enables qemu emulators for building Docker images for arm64/armv6/armv7/etc on the host.
docker run --privileged --rm tonistiigi/binfmt --install all

dagger run go run ./cmd \
  package \
  --distro=linux/amd64 \
  --distro=linux/arm64 \
  --build-id=${DRONE_BUILD_NUMBER} \
  --grafana-dir=${GRAFANA_DIR} \
  --github-token=${GITHUB_TOKEN} \
  --go-version=${GO_VERSION} \
  --destination=${DESTINATION}/${DRONE_BUILD_EVENT} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64}
