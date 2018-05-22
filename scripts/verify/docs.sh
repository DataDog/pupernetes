#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -e

cd $(dirname $0)/../..

make docs

DIFF=$(git diff docs/)
if [[ "${DIFF}x" != "x" ]]
then
    echo "Docs outdated:" >&2
    echo ${DIFF} >&2
    exit 2
fi

DIFF=$(git ls-files docs/ --exclude-standard --others)
if [[ "${DIFF}x" != "x" ]]
then
    echo "Docs removed:" >&2
    echo ${DIFF} >&2
    exit 2
fi

exit 0
