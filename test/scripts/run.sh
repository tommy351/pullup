#!/bin/bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." ; pwd)"
PATH="${PROJECT_ROOT}/assets/bin:${PATH}"

kustomize build "$PROJECT_ROOT/test/deployment" | kubectl apply -f -
