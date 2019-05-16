#!/bin/bash

set -euo pipefail

go mod vendor
docker-compose run --rm codegen
