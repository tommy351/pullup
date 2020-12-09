#!/bin/bash

set -euo pipefail

export JUNIT_OUTPUT="${PWD}/reports/junit"

go run github.com/onsi/ginkgo/ginkgo \
  -r \
  --randomizeAllSpecs \
  --randomizeSuites \
  --failOnPending \
  --cover \
  --trace \
  --race \
  --progress \
  --skipPackage e2e \
  "$@"

$(dirname ${BASH_SOURCE[0]})/collect-coverage.sh
