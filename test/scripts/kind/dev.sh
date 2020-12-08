#!/bin/bash

set -euo pipefail

NAMESPACE=test-pullup
RESOURCES_TO_DELETE=(pod/test-pullup-e2e webhook/test-http-server)

# Build images
$(dirname ${BASH_SOURCE[0]})/../build-images.sh

# Setup the cluster
source $(dirname ${BASH_SOURCE[0]})/setup.sh

# Delete resources
for name in "${RESOURCES_TO_DELETE[@]}"
do
  kubectl delete "$name" -n "$NAMESPACE" --ignore-not-found=true
done

# Run tests
$(dirname ${BASH_SOURCE[0]})/../run.sh
