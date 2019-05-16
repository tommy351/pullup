#!/bin/bash

set -euo pipefail

find $(dirname "${BASH_SOURCE[0]}")/.. -type f -name *.coverprofile -exec cat {} \; > coverage.txt
