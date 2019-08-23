#!/bin/bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/../../.." ; pwd)"
export KIND_CLUSTER_NAME='pullup-test'
export PATH="${PROJECT_ROOT}/assets/bin:${PATH}"
