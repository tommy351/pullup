#!/bin/bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." ; pwd)"
PATH="${PROJECT_ROOT}/assets/bin:${PATH}"
CRDS=(
  webhooks
  resourcesets
  resourcetemplates
  httpwebhooks
  githubwebhooks
)

# Create CRDs first
kubectl apply -f "${PROJECT_ROOT}/deployment/base/crds"

# Wait until CRDs are established
for crd in "${CRDS[@]}"
do
  kubectl wait --for=condition=established --timeout=60s "crd/${crd}.pullup.dev"
done
