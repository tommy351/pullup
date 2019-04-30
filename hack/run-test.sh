#!/bin/bash

set -euxo pipefail

export JUNIT_OUTPUT="${PWD}/reports/junit"

go get github.com/onsi/ginkgo/ginkgo
ginkgo -r --randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --race --progress $@
./hack/collect-coverage.sh
