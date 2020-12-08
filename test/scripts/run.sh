#!/bin/bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." ; pwd)"
PATH="${PROJECT_ROOT}/assets/bin:${PATH}"
NAMESPACE=test-pullup
JOB_NAME=pullup-e2e

$(dirname ${BASH_SOURCE[0]})/create-crd.sh

# Apply rest of the manifests
kubectl apply -k "${PROJECT_ROOT}/test/deployment"

# Wait until the job is running
until kubectl get pod -l "job-name=${JOB_NAME}" -n "$NAMESPACE" | grep Running
do
  sleep 1
done

# Print job logs
kubectl logs -n "$NAMESPACE" -f "job/${JOB_NAME}"

# Wait until the job completed
kubectl wait -n "$NAMESPACE" --for=condition=complete --timeout=5s "job/${JOB_NAME}"
