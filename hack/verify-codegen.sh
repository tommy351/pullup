#!/bin/bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/.." ; pwd)"
DIFFROOT=${PROJECT_ROOT}
TMP_DIFFROOT=$(mktemp -d)

cleanup() {
  echo "Deleting ${TMP_DIFFROOT}"
  rm -rf ${TMP_DIFFROOT}
}
trap "cleanup" EXIT SIGINT

echo "Copying from ${DIFFROOT} to ${TMP_DIFFROOT}"
cp -a ${DIFFROOT}/* ${TMP_DIFFROOT}
${PROJECT_ROOT}/hack/update-codegen.sh

echo "Diffing ${DIFFROOT} against freshly generated codegen"
ret=0
diff -x '.*' -qr ${DIFFROOT} ${TMP_DIFFROOT} || ret=$?

echo "Restoring ${DIFFROOT}"
cp -a ${TMP_DIFFROOT}/* ${DIFFROOT}

if [[ $ret -eq 0 ]]
then
  echo "${DIFFROOT} up to date."
else
  echo "${DIFFROOT} is out of date. Please run hack/update-codegen.sh"
  exit $ret
fi
