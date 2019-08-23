#!/bin/bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." ; pwd)"
PATH="${PROJECT_ROOT}/assets/bin:${PATH}"

build_deployment() {
  kustomize build "$PROJECT_ROOT/test/deployment"
}

build_deployment | kubectl delete -f - || true
build_deployment | kubectl apply -f -

kubectl wait --for=condition=complete --timeout=120s job/pullup-test-e2e
