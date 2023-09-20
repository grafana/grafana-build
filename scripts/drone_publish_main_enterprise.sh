#!/usr/bin/env sh

dagger run go run ./cmd \
  package \
  --distro=linux/amd64 \
  --distro=linux/arm64 \
  --enterprise \
  --grafana=false \
  --build-id=${DRONE_BUILD_NUMBER} \
  --grafana-dir=${GRAFANA_DIR} \
  --github-token=${GITHUB_TOKEN} \
  --go-version=${GO_VERSION} \
  --destination=${DESTINATION}/${DRONE_BUILD_EVENT} \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64}
