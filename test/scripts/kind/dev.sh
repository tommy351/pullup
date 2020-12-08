#!/bin/bash

set -euo pipefail

NAMESPACE=test-pullup
RESOURCES_TO_DELETE=(
  deployment/test-pullup-controller
  deployment/test-pullup-webhook
  pod/test-pullup-e2e
  webhook/test-http-server
  httpwebhook/test-http-server
)

# Build images
$(dirname ${BASH_SOURCE[0]})/../build-images.sh

# Setup the cluster
source $(dirname ${BASH_SOURCE[0]})/setup.sh

# Create CRDs
$(dirname ${BASH_SOURCE[0]})/../create-crd.sh

# Delete resources
for name in "${RESOURCES_TO_DELETE[@]}"
do
  kubectl delete "$name" -n "$NAMESPACE" --ignore-not-found=true
done

# Run tests
$(dirname ${BASH_SOURCE[0]})/../run.sh
