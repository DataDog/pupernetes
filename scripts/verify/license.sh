#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -e

cd $(dirname $0)/../..

make gen-license

DIFF=$(git diff LICENSE-3rdparty.csv)
if [[ "${DIFF}x" != "x" ]]
then
    echo "License outdated:" >&2
    echo ${DIFF} >&2
    exit 2
fi

DIFF=$(git ls-files docs/ --exclude-standard --others)
if [[ "${DIFF}x" != "x" ]]
then
    echo "License removed:" >&2
    echo ${DIFF} >&2
    exit 2
fi

exit 0
