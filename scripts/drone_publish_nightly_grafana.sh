#!/usr/bin/env bash
set -e
ver="nightly-${DRONE_COMMIT_SHA:0:8}"
local_dir="${DRONE_WORKSPACE}/dist"

# Publish the docker images present in the bucket
dagger run --silent go run ./cmd docker publish \
  $(find $local_dir | grep docker.tar.gz | grep -v sha256 | grep -v enterprise | awk '{print "--package=file://"$0}') \
  --username=${DOCKER_USERNAME} \
  --password=${DOCKER_PASSWORD} \
  --repo="grafana-dev"

# Copy only the linux/amd64 edition storybook into a separate folder
dagger run --silent go run ./cmd storybook \
  $(find $local_dir | grep tar.gz | grep linux | grep amd64 | grep -v sha256 | awk '{print "--package=file://"$0}') \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} \
  --destination="${STORYBOOK_DESTINATION}/${ver}"

# Copy only the linux/amd64 edition static assets into a separate folder
dagger run --silent go run ./cmd cdn \
  $(find $local_dir | grep tar.gz | grep linux | grep amd64 | grep -v sha256 | awk '{print "--package=file://"$0}') \
  --gcp-service-account-key-base64=${GCP_KEY_BASE64} \
  --destination="${CDN_DESTINATION}/${ver}/public"