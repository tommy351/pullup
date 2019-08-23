#!/bin/bash

set -euo pipefail

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
build_image pullup-test-e2e "${PROJECT_ROOT}/test/e2e"
