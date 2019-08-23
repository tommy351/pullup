#!/bin/bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." ; pwd)"
BUILD_CONTEXT=$PROJECT_ROOT/test/bin

function build_image {
  local name=$1

  GOOS=linux CGO_ENABLED=0 go build -o "${BUILD_CONTEXT}/${name}" "${PROJECT_ROOT}/cmd/${name}"
  docker build \
    -t "tommy351/pullup-${name}" \
    --build-arg "BINARY_NAME=${name}" \
    -f "${PROJECT_ROOT}/Dockerfile" \
    "$BUILD_CONTEXT"
}

mkdir -p "$BUILD_CONTEXT"
build_image controller
build_image webhook
