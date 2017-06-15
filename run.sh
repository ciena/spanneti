#!/bin/bash

docker run --rm -ti \
  --pid=host \
  --security-opt apparmor:unconfined \
  --cap-add=NET_ADMIN \
  --cap-add=SYS_ADMIN \
  --cap-add=SYS_PTRACE \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $GOPATH/src:/go/src \
  golang:1.8.1 \
  bash -c 'cd /go/src/bitbucket.ciena.com/BP_ONOS/spanneti; bash'
