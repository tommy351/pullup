#!/bin/bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/.." ; pwd)"
DEPLOYMENT_DIR=$PROJECT_ROOT/deployment
TMP_DEPLOYMENT_DIR=$(mktemp -d)
IMAGE_TAG=${1:-latest}

cleanup() {
  rm -rf "$TMP_DEPLOYMENT_DIR"
}
trap "cleanup" EXIT SIGINT

cp -a "${DEPLOYMENT_DIR}"/* "$TMP_DEPLOYMENT_DIR"

echo "
images:
  - name: tommy351/pullup-controller
    newTag: ${IMAGE_TAG}
  - name: tommy351/pullup-webhook
    newTag: ${IMAGE_TAG}
" >> "${TMP_DEPLOYMENT_DIR}/kustomization.yml"

"$PROJECT_ROOT/assets/bin/kubectl" kustomize "$TMP_DEPLOYMENT_DIR"
