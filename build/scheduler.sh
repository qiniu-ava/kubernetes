#!/bin/bash

# example: ./build/scheduler.sh
# example: ./build/scheduler.sh --push

set -ex

push=false
tag=false
if [ "$1" == "--push" ]; then
	push=true
fi

if [ "$2" == "--tag" ]; then
	tag=true
fi

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd ${DIR}/.. # project root path

./build/run.sh make WHAT=plugin/cmd/kube-scheduler

mkdir -p ./_output/images/kube-scheduler
cp ./build/build-image/ava-scheduler.Dockerfile ./_output/images/kube-scheduler/ava-scheduler.Dockerfile
docker build -t ava-kube-scheduler:latest -f ./_output/images/kube-scheduler/ava-scheduler.Dockerfile ./_output/dockerized

if $push; then
		VERSION=$(date -u '+%Y%m%d')-$(git rev-parse --short HEAD)
		docker tag ava-kube-scheduler:latest reg-xs.qiniu.io/atlab/ava-kube-scheduler:$VERSION
	echo "pushing reg-xs.qiniu.io/atlab/ava-kube-scheduler:$VERSION"
		docker push reg-xs.qiniu.io/atlab/ava-kube-scheduler:$VERSION
fi
