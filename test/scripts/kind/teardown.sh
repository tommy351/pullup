#!/bin/bash

set -euo pipefail

source $(dirname ${BASH_SOURCE[0]})/base.sh

kind delete cluster --name "$KIND_CLUSTER_NAME"
