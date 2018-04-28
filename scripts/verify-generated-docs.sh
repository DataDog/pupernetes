#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

cd $(dirname $0)/..

make gen-docs
DIFF=$(git diff docs/)

if [[ "${DIFF}x" == "x" ]]
then
    exit 0
fi

echo "Docs outdated:" >&2
echo ${DIFF} >&2

exit 2
