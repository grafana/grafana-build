#!/usr/bin/env bash
set -e
local_dir="${DRONE_WORKSPACE}/dist"

# Publish the docker images present in the bucket
dagger run --silent go run ./cmd docker publish \
  $(find $local_dir | grep docker.tar.gz | grep -v sha256 | grep -v enterprise | awk '{print "--package=file://"$0}') \
  --username=${DOCKER_USERNAME} \
  --password=${DOCKER_PASSWORD} \
  --repo="grafana-nightly"