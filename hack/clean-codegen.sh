#!/bin/bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/.." ; pwd)"

rm -rf \
  "${PROJECT_ROOT}/deployment/base/crd" \
  "${PROJECT_ROOT}/deployment/base/rbac" \
  "${PROJECT_ROOT}/pkg/apis/pullup/*/zz_generated.deepcopy.go"
