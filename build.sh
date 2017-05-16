#!/bin/bash

# run a build container, with static compilation
docker run --rm \
  -v $GOPATH:/go \
  golang:1.8.1 \
  bash -c "cd /go/src/github.com/khagerma/cord-networking; go build -v --ldflags '-extldflags \"-static\"' -o build/cord-network-manager main.go"

docker build -t cord-network-manager .

if [ "$1" == "--run" ]; then
	docker run --rm \
    --pid=host \
    --security-opt apparmor:unconfined \
    --cap-add=NET_ADMIN \
    --cap-add=SYS_ADMIN \
    --cap-add=SYS_PTRACE \
    -v /var/run/docker.sock:/var/run/docker.sock \
    cord-network-manager
fi