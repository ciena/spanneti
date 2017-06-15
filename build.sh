#!/bin/bash

# run a build container, with static compilation
docker run --rm \
  -v $GOPATH/src:/go/src \
  golang:1.8.1 \
  bash -c "cd /go/src/bitbucket.ciena.com/BP_ONOS/spanneti; go build -v -tags netgo --ldflags '-extldflags \"-static\"' -o build/spanneti main.go"

docker build -t spanneti .

if [ "$1" == "--run" ]; then
	docker run --rm \
    --pid=host \
    --security-opt apparmor:unconfined \
    --cap-add=NET_ADMIN \
    --cap-add=SYS_ADMIN \
    --cap-add=SYS_PTRACE \
    -v /var/run/docker.sock:/var/run/docker.sock \
    spanneti
fi