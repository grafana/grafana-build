#!/usr/bin/env sh

dagger run go run ./cmd \
  package \
  --distro=linux/amd64 \
  --distro=linux/arm64 \
  --build-id=${DRONE_BUILD_NUMBER} \
  --grafana-dir=${GRAFANA_DIR} \
  --github-token=${GITHUB_TOKEN} \
  --destination=${DESTINATION}/${DRONE_BUILD_EVENT} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64}
