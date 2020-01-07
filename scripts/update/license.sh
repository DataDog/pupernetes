#!/usr/bin/env bash

set -exo pipefail

export LC_ALL=C
ROOT=$(git rev-parse --show-toplevel)

cd $(dirname $0)/../..

$ROOT/bin/wwhrd list

echo Component,Origin,License > LICENSE-3rdparty.csv
echo 'core,"github.com/frapposelli/wwhrd",MIT' >> LICENSE-3rdparty.csv

$ROOT/bin/wwhrd list |& grep "Found License" | awk '{print $5,$4}' | sed -r "s/\x1B\[([0-9]{1,2}(;[0-9]{1,2})?)?[mGK]//g" | sed s/" license="/,/ | sed s/package=/core,/ | sort >> LICENSE-3rdparty.csv
