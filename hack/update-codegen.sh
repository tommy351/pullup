#!/bin/bash

set -euxo pipefail

go mod vendor
docker-compose run --rm codegen
