#!/usr/bin/env sh

go run ./cmd \
  --build-id=${DRONE_BUILD_ID} \
  --grafana-dir=${GRAFANA_DIR} \
  --github-token=${GITHUB_TOKEN} \
  package \
  --distro=${DISTROS} \
  publish \
  --destination=${DESTINATION}/${DRONE_BUILD_EVENT} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64}
