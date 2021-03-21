#!/bin/bash

set -euo pipefail

export GO111MODULE="on"

PROJECT_ROOT="$(cd "$(dirname "$0")/.." ; pwd)"

echo "Generate CRD"
go run sigs.k8s.io/controller-tools/cmd/controller-gen \
  crd:trivialVersions=true \
  object:headerFile="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  paths="${PROJECT_ROOT}/pkg/apis/..." \
  output:crd:artifacts:config="${PROJECT_ROOT}/deployment/base/crds"

echo "Generate RBAC"
go run sigs.k8s.io/controller-tools/cmd/controller-gen \
  rbac:roleName=pullup \
  object:headerFile="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  paths="${PROJECT_ROOT}/cmd/...;${PROJECT_ROOT}/internal/..." \
  output:rbac:artifacts:config="${PROJECT_ROOT}/deployment/base/rbac"

go get github.com/google/wire/cmd/wire@v0.4.0

echo "Generate Go files"
go generate ./...

echo "Removing unused modules"
go mod tidy
