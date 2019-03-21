DOCKER_COMPOSE := docker-compose

# find local go if exists and fallback to docker if not
GOEXEC = $(shell which go)
GO_DOCKER = ${DOCKER_COMPOSE} run --rm -e GOOS=linux go
ifeq (${GOEXEC},)
GOEXEC = ${GO_DOCKER}
endif

# find local go-bindata if exists and fallback to docker if not
GOBINDATA_EXISTS = $(shell test -f ${GOPATH}/bin/go-bindata && echo "1" || echo "0")
ifeq (${GOBINDATA_EXISTS},1)
GOBINDATA = ${GOPATH}/bin/go-bindata
else
GOBINDATA = ${DOCKER_COMPOSE} run --rm --entrypoint=/go/bin/go-bindata go
endif

build: build-schema-assets
	${GOEXEC} build ./...

build-schema-assets:
	${GOBINDATA} -pkg schema -o schema/schemas.go --prefix "schemas/" schemas/...

test: build-schema-assets
	${GOEXEC} test ./...
