#!/bin/bash

set -euo pipefail

export DOCKER_BUILDKIT=1

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." ; pwd)"
CMD_DIR="${PROJECT_ROOT}/cmd"
BIN_DIR="${PROJECT_ROOT}/test/bin"

function build_image {
  local name=$1
  local pkg_path=$2
  local output_path="${BIN_DIR}/${name}"

  GOOS=linux CGO_ENABLED=0 go build -o "$output_path" "$pkg_path"
  docker build \
    -t "$name" \
    --build-arg "BINARY_NAME=${name}" \
    -f "${PROJECT_ROOT}/Dockerfile" \
    "$BIN_DIR"
}

mkdir -p "$BIN_DIR"
build_image pullup-controller "${CMD_DIR}/controller"
build_image pullup-webhook "${CMD_DIR}/webhook"
build_image pullup-http-server "${PROJECT_ROOT}/test/http-server"

go get github.com/onsi/ginkgo/ginkgo
GOOS=linux CGO_ENABLED=0 ginkgo build "${PROJECT_ROOT}/test/e2e"
docker build \
  -t pullup-e2e \
  --build-arg "BINARY_NAME=e2e.test" \
  -f "${PROJECT_ROOT}/Dockerfile" \
  "${PROJECT_ROOT}/test/e2e"
