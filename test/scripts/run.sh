#!/bin/bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." ; pwd)"
PATH="${PROJECT_ROOT}/assets/bin:${PATH}"
CRDS=(webhooks.pullup.dev resourcesets.pullup.dev)
NAMESPACE=test-pullup

# Create CRDs first
kubectl apply -f "${PROJECT_ROOT}/deployment/base/crds"

# Wait until CRDs are established
for crd in "${CRDS[@]}"
do
  kubectl wait --for=condition=established --timeout=60s "crd/${crd}"
done

# Apply rest of the manifests
kustomize build "${PROJECT_ROOT}/test/deployment" | kubectl apply -f -

# Wait until deployments are available
for deploy in $(kubectl get deploy -o name)
do
  kubectl wait -n "$NAMESPACE" --for=condition=available --timeout=60s "$deploy"
done

# Wait until the job completed
go run "${PROJECT_ROOT}/test/scripts/wait-for-job" \
  --job test-pullup-e2e \
  --namespace "$NAMESPACE"
