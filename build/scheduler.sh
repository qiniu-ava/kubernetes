#!/bin/bash

# example: ./build/scheduler.sh
# example: ./build/scheduler.sh --push

set -ex

push=false
if [ "$1" == "--push" ]; then
    push=true
fi

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd ${DIR}/..     # project root path

mkdir -p _output/bin
touch _output/bin/deepcopy-gen
touch _output/bin/conversion-gen
touch _output/bin/defaulter-gen
touch _output/bin/openapi-gen
chmod u=rwx _output/bin/deepcopy-gen
chmod u=rwx _output/bin/conversion-gen
chmod u=rwx _output/bin/defaulter-gen
chmod u=rwx _output/bin/openapi-gen

KUBE_BUILD_PLATFORMS=linux/amd64 make WHAT=plugin/cmd/kube-scheduler
docker build -t ava-kube-scheduler:latest -f ./build/build-image/ava-scheduler.Dockerfile .

if $push; then
    docker tag ava-kube-scheduler:latest reg-xs.qiniu.io/atlab/ava-kube-scheduler:latest
    docker push reg-xs.qiniu.io/atlab/ava-kube-scheduler:latest
fi