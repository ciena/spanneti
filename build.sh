#!/bin/bash

GIT_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
GIT_COMMIT_NUM="$(git rev-list --count HEAD)"
GIT_COMMIT="$(git log --format='%H' -n 1)"
if [ "$(git ls-files --others --modified --exclude-standard | grep '.*\.go')" != "" ]; then
	CHANGED="true"
fi

# run a build container, with static compilation
docker run --rm \
  -v $GOPATH/src:/go/src \
  golang:1.8.1 \
  bash -c "cd /go/src/bitbucket.ciena.com/BP_ONOS/spanneti; go build -v -tags netgo --ldflags '-extldflags \"-static\" -X \"main.GIT_BRANCH=$GIT_BRANCH\" -X \"main.GIT_COMMIT_NUM=$GIT_COMMIT_NUM\" -X \"main.GIT_COMMIT=$GIT_COMMIT\" -X \"main.CHANGED=$CHANGED\"' -o build/spanneti main.go"

docker build -t spanneti .
if [ "$CHANGED" != "true" ]; then
	if [ "$GIT_BRANCH" == "master" ]; then
		docker tag spanneti "khagerma/spanneti:0.$GIT_COMMIT_NUM"
	else
		docker tag spanneti "khagerma/spanneti:$GIT_BRANCH-0.$GIT_COMMIT_NUM"
	fi
fi

IMAGES="$(docker images -q --filter=dangling=true)"
if [ "$IMAGES" != "" ]; then
	docker rmi $IMAGES
fi

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