#!/bin/bash

# example: ./build/scheduler.sh
# example: ./build/scheduler.sh --push

set -ex

push=false
if [ "$1" == "--push" ]; then
	push=true
fi

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd ${DIR}/.. # project root path

./build/run.sh make WHAT=plugin/cmd/kube-scheduler

mkdir -p ./_output/images/kube-scheduler
cp ./build/build-image/ava-scheduler.Dockerfile ./_output/images/kube-scheduler/ava-scheduler.Dockerfile
docker build -t ava-kube-scheduler:latest -f ./_output/images/kube-scheduler/ava-scheduler.Dockerfile ./_output/dockerized

if $push; then
	docker tag ava-kube-scheduler:latest reg-xs.qiniu.io/atlab/ava-kube-scheduler:latest
	docker push reg-xs.qiniu.io/atlab/ava-kube-scheduler:latest
fi
