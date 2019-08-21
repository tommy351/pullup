#!/bin/bash

set -euo pipefail

go get k8s.io/code-generator/cmd/{defaulter-gen,client-gen,lister-gen,informer-gen,deepcopy-gen}

PROJECT_ROOT="$(cd "$(dirname "$0")/.." ; pwd)"
TMP_DIR=$(mktemp -d)

cleanup() {
  rm -rf ${TMP_DIR}
}
trap "cleanup" EXIT SIGINT

${PROJECT_ROOT}/hack/generate-groups.sh all \
  github.com/tommy351/pullup/pkg/client \
  github.com/tommy351/pullup/pkg/apis \
  pullup:v1alpha1 \
  --output-base ${TMP_DIR} \
  --go-header-file ${PROJECT_ROOT}/hack/boilerplate.go.txt

echo "Copying generated file to ${PROJECT_ROOT}/pkg"
cp -a ${TMP_DIR}/github.com/tommy351/pullup/pkg/* ${PROJECT_ROOT}/pkg

go get github.com/google/wire/cmd/wire

echo "Generate Go files"
go generate ./...

echo "Removing unused modules"
go mod tidy
