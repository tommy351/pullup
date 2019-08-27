#!/bin/bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." ; pwd)"
PATH="${PROJECT_ROOT}/assets/bin:${PATH}"
CRDS=(webhooks.pullup.dev resourcesets.pullup.dev)
NAMESPACE=test-pullup
JOB_NAME=test-pullup-e2e

# Create CRDs first
kubectl apply -f "${PROJECT_ROOT}/deployment/base/crds"

# Wait until CRDs are established
for crd in "${CRDS[@]}"
do
  kubectl wait --for=condition=established --timeout=60s "crd/${crd}"
done

# Apply rest of the manifests
kustomize build "${PROJECT_ROOT}/test/deployment" | kubectl apply -f -

# Wait until the job is running
until kubectl get pod -l "job-name=${JOB_NAME}" -n "$NAMESPACE" | grep Running
do
  sleep 1
done

# Print job logs
kubectl logs -n "$NAMESPACE" -f "job/${JOB_NAME}"

# Wait until the job completed
kubectl wait -n "$NAMESPACE" --for=condition=complete --timeout=5s "job/${JOB_NAME}"
