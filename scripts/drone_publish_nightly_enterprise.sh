#!/usr/bin/env bash
set -e
ver=$(cat ${GRAFANA_DIR}/package.json | jq -r .version)
local_dir="${DRONE_WORKSPACE}/dist"

# Check if version has hyphen
if [[ $ver == *-* ]]; then
    # If it does, replace everything after the hyphen
    ver=$(echo $ver | sed -E "s/-.*/-nightly.${DRONE_COMMIT_SHA:0:8}/")
else
    # If it doesn't, append "-nightly.${DRONE_COMMIT_SHA:0:8}"
    ver="${ver}-nightly.${DRONE_COMMIT_SHA:0:8}"
fi

# Publish the docker images present in the bucket
dagger run --silent go run ./cmd docker publish \
  $(find $local_dir | grep docker.tar.gz | grep -v sha256 | awk '{print "--package=file://"$0}') \
  --username=${DOCKER_USERNAME} \
  --password=${DOCKER_PASSWORD} \
  --repo="grafana-enterprise-dev"

# Publish packages to the downloads bucket
dagger run --silent go run ./cmd package publish \
  $(find $local_dir | grep -e .rpm -e .tar.gz -e .exe -e .zip -e .deb | awk '{print "--package=file://"$0}') \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} \
  --destination="${DOWNLOADS_DESTINATION}/enterprise/release"

# Publish packages to grafana.com
dagger run --silent go run ./cmd gcom publish \
  $(find $local_dir | grep -e .rpm -e .tar.gz -e .exe -e .zip -e .deb | grep -v sha256 | awk '{print "--package=file://"$0}') \
  --api-key=${GCOM_API_KEY} \
  --api-url="https://grafana-dev.com/api/grafana-enterprise" \
  --download-url="https://dl.grafana.com/enterprise/release"