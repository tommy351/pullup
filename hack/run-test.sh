#!/bin/bash

set -euo pipefail

export JUNIT_OUTPUT="${PWD}/reports/junit"

go get github.com/onsi/ginkgo/ginkgo
ginkgo -r --randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --race --progress $@
$(dirname ${BASH_SOURCE[0]})/collect-coverage.sh
