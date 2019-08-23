#!/bin/bash

set -euo pipefail

source $(dirname ${BASH_SOURCE[0]})/setup.sh

cleanup() {
  $(dirname ${BASH_SOURCE[0]})/teardown.sh
}
trap "cleanup" EXIT SIGINT

$(dirname ${BASH_SOURCE[0]})/../run.sh
