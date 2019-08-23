#!/bin/bash

set -euo pipefail

source $(dirname ${BASH_SOURCE[0]})/base.sh

TEST_IMAGES=(pullup-controller pullup-webhook pullup-test-e2e)

# Create kind cluster if not exists
if ! kind get clusters | grep -lq "$KIND_CLUSTER_NAME"
then
  kind create cluster --name "$KIND_CLUSTER_NAME"
fi

# Set KUBECONFIG for kubectl
export KUBECONFIG=$(kind get kubeconfig-path --name "$KIND_CLUSTER_NAME")

# Load images
for img in "${TEST_IMAGES[@]}"
do
  echo "Loading Docker image: ${img}"
  kind load docker-image --name "$KIND_CLUSTER_NAME" "$img"
done
