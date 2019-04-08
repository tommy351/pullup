#!/bin/bash

set -euxo pipefail

find . -type f -name *.coverprofile -exec cat {} \; > coverage.txt
