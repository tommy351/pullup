#!/bin/bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." ; pwd)"
PATH="${PROJECT_ROOT}/assets/bin:${PATH}"
CRDS=(webhooks resourcesets resourcetemplates httpwebhooks githubwebhooks)
NAMESPACE=test-pullup
POD_NAME=test-pullup-e2e

# Create CRDs first
kubectl apply -f "${PROJECT_ROOT}/deployment/base/crds"

# Wait until CRDs are established
for crd in "${CRDS[@]}"
do
  kubectl wait --for=condition=established --timeout=60s "crd/${crd}.pullup.dev"
done

# Apply rest of the manifests
kubectl apply -k "${PROJECT_ROOT}/test/deployment"

# Wait until the pod is ready
kubectl wait --for=condition=Ready --timeout=60s "pod/${POD_NAME}" -n "$NAMESPACE"

# Print job logs
kubectl logs -n "$NAMESPACE" -f "pod/${POD_NAME}"
