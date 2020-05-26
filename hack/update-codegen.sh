#!/bin/bash

set -euo pipefail

export GO111MODULE="on"

PROJECT_ROOT="$(cd "$(dirname "$0")/.." ; pwd)"

go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0

echo "Generate CRD"
controller-gen \
  crd:trivialVersions=true \
  object:headerFile="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  paths="${PROJECT_ROOT}/pkg/apis/..." \
  output:crd:artifacts:config="${PROJECT_ROOT}/deployment/base/crds"

go get k8s.io/code-generator/cmd/{defaulter-gen,client-gen,lister-gen,informer-gen,deepcopy-gen}@v0.18.2

TMP_DIR=$(mktemp -d)

cleanup() {
  rm -rf "$TMP_DIR"
}
trap "cleanup" EXIT SIGINT

"${PROJECT_ROOT}/hack/generate-groups.sh" client,lister,informer \
  github.com/tommy351/pullup/pkg/client \
  github.com/tommy351/pullup/pkg/apis \
  pullup:v1alpha1 \
  --output-base "$TMP_DIR" \
  --go-header-file "${PROJECT_ROOT}/hack/boilerplate.go.txt"

echo "Copying generated file to ${PROJECT_ROOT}/pkg"
cp -a "$TMP_DIR"/github.com/tommy351/pullup/pkg/* "${PROJECT_ROOT}/pkg"

go get github.com/google/wire/cmd/wire@v0.4.0

echo "Generate Go files"
go generate ./...

echo "Removing unused modules"
go mod tidy
