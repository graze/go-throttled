DOCKER_COMPOSE := docker-compose

# find local go if exists and fallback to docker if not
GOEXEC = $(shell which go)
GO_DOCKER = ${DOCKER_COMPOSE} run --rm -e GOOS=linux go
ifeq (${GOEXEC},)
GOEXEC = ${GO_DOCKER}
endif

default: get build test

get:
	${GOEXEC} get ${gobuild_args} ./...

build:
	${GOEXEC} build ./...

test:
	${GOEXEC} test ./...
