#!/bin/bash

docker rmi spanneti

GIT_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
GIT_COMMIT_NUM="$(git rev-list --count HEAD)"
GIT_COMMIT="$(git log --format='%H' -n 1)"
if [ "$(git ls-files --others --modified --exclude-standard | grep '.*\.go')" != "" ]; then
	CHANGED="true"
fi

docker build -t spanneti \
	--build-arg "GIT_BRANCH=$GIT_BRANCH" \
	--build-arg "GIT_COMMIT_NUM=$GIT_COMMIT_NUM" \
	--build-arg "GIT_COMMIT=$GIT_COMMIT" \
	--build-arg "CHANGED=$CHANGED" .

if [ "$CHANGED" != "true" ]; then
	if [ "$GIT_BRANCH" == "master" ]; then
		docker rmi "khagerma/spanneti:0.$GIT_COMMIT_NUM" || true
		docker tag spanneti "khagerma/spanneti:0.$GIT_COMMIT_NUM"
	else
		docker rmi "khagerma/spanneti:$(echo "$GIT_BRANCH" | sed -e 's/\//_/g')-0.$GIT_COMMIT_NUM" || true
		docker tag spanneti "khagerma/spanneti:$(echo "$GIT_BRANCH" | sed -e 's/\//_/g')-0.$GIT_COMMIT_NUM"
	fi
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