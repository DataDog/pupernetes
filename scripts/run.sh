#!/bin/bash

set -o pipefail
set -ex

apt-get update -qq
apt-get install -y make curl unzip

curl https://get.docker.com | sh
