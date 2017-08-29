FROM golang:1.9.0 as build

WORKDIR /go/src/bitbucket.ciena.com/BP_ONOS/spanneti

# get and install dependencies
RUN go get -u github.com/kardianos/govendor
COPY vendor/vendor.json vendor/
RUN govendor sync -v
RUN govendor install -tags netgo +vendor

# get git status
ARG GIT_BRANCH
ARG GIT_COMMIT_NUM
ARG GIT_COMMIT
ARG CHANGED

# bring in source files
COPY . ./

# build static binary
RUN go build -v -tags netgo \
	--ldflags "-extldflags \"-static\" \
	-X \"main.GIT_BRANCH=$GIT_BRANCH\" \
	-X \"main.GIT_COMMIT_NUM=$GIT_COMMIT_NUM\" \
	-X \"main.GIT_COMMIT=$GIT_COMMIT\" \
	-X \"main.CHANGED=$CHANGED\"" \
	-o build/spanneti main.go

# create final container
FROM alpine:3.5
COPY --from=build /go/src/bitbucket.ciena.com/BP_ONOS/spanneti/build/spanneti /bin/spanneti
CMD ["spanneti"]
