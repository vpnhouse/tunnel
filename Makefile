GIT_TAG    ?= $(shell git describe --tags --abbrev=0)
GIT_COMMIT ?= $(shell git rev-parse --short HEAD)
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
DOCKER_TAG ?= $(shell git describe --tags --exact-match `git rev-parse --short HEAD` 2>/dev/null || git rev-parse --abbrev-ref HEAD | sed -e 's/\//-/g')

ifeq (${GIT_COMMIT},)
	GIT_COMMIT = unknown
endif

ifeq (${GIT_TAG},)
	GIT_TAG = unknown
endif

ifeq (${DOCKER_TAG},)
	DOCKER_TAG = unknown
endif

DEFAULT_FEATURE_SET=personal
FEATURE_SET  ?= $(DEFAULT_FEATURE_SET)
ifneq ($(FEATURE_SET), $(DEFAULT_FEATURE_SET))
DOCKER_TAG := $(DOCKER_TAG)-$(FEATURE_SET)
endif
DOCKER_IMAGE ?= vpnhouse/tunnel:$(DOCKER_TAG)

GO_VERSION_PATH = github.com/vpnhouse/tunnel/pkg/version
GO_LDFLAGS = -w -s -X $(GO_VERSION_PATH).tag=$(GIT_TAG) -X $(GO_VERSION_PATH).commit=$(GIT_COMMIT) -X $(GO_VERSION_PATH).feature=$(FEATURE_SET)
DOCKER_BUILD_ARGS = --progress=plain --platform=linux/amd64 --build-arg FEATURE_SET=$(FEATURE_SET) --file ./Dockerfile
DESCRIPTION = tunnel $(GIT_TAG)-$(GIT_COMMIT) branch=$(GIT_BRANCH) features=$(FEATURE_SET)
GO_FLAGS=-trimpath

ifeq ($(FEATURE_SET),enterprise)
    GO_FLAGS+=-tags iprose
endif

run:
	@echo "+ $@ $(DESCRIPTION)"
	go run ./cmd/tunnel/main.go

# its important to build the frontend first
# because the app will embed it into itself
build/all: build/frontend build/app

build/app:
	@echo "+ $@ $(DESCRIPTION)"
	go build -ldflags="$(GO_LDFLAGS)" $(GO_FLAGS) -o tunnel-node ./cmd/tunnel/main.go

build/frontend:
	@echo "+ $@ $(DESCRIPTION)"
	rm -rf ./frontend/dist
	rm -rf ./internal/frontend/dist && mkdir ./internal/frontend/dist
	touch ../tunnel/internal/frontend/dist/stub.html
	cd ./frontend && npm run build
	cp -r ./frontend/dist/* ./internal/frontend/dist/


docker/all: docker/build docker/push

docker/build:
	@echo "+ $@ $(DOCKER_IMAGE)"
	DOCKER_BUILDKIT=1 docker build --ssh default $(DOCKER_BUILD_ARGS) --tag $(DOCKER_IMAGE) .

docker/push:
	@echo "+ $@ $(DOCKER_IMAGE)"
	@docker push $(DOCKER_IMAGE)

test:
	@echo "+ $@"
	go test -v -race ./...

fmt:
	@echo "+ $@"
	@test -z "$$(gofmt -s -l . 2>&1 | grep -v ^vendor/ | tee /dev/stderr)" || \
		(echo >&2 "+ please format Go code with 'gofmt -s'" && false)

vet:
	@echo "+ $@"
	@go vet ./...

.PHONY: proto
proto:
	@protoc -I proto/ --go_out=./proto/ --go-grpc_out=./proto/ proto/*.proto
	@mv ./proto/github.com/vpnhouse/tunnel/proto/*.pb.go ./proto
	@rm -rf ./proto/github.com

version:
	@echo "+ $@ $(DESCRIPTION)"

