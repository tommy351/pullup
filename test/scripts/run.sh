#!/bin/bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." ; pwd)"
PATH="${PROJECT_ROOT}/assets/bin:${PATH}"
NAMESPACE=test-pullup
POD_NAME=pullup-e2e

$(dirname ${BASH_SOURCE[0]})/create-crd.sh

# Apply rest of the manifests
kubectl apply -k "${PROJECT_ROOT}/test/deployment"

# Wait until the pod is ready
kubectl wait --for=condition=Ready --timeout=60s "pod/${POD_NAME}" -n "$NAMESPACE"

# Print job logs
kubectl logs -n "$NAMESPACE" -f "pod/${POD_NAME}"
