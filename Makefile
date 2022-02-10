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

DESCRIPTION = tunnel $(GIT_TAG)-$(GIT_COMMIT) branch $(GIT_BRANCH)
GO_VERSION_PATH = github.com/Codename-Uranium/tunnel/pkg/version
GO_LDFLAGS = -w -s -X $(GO_VERSION_PATH).tag=$(GIT_TAG) -X $(GO_VERSION_PATH).commit=$(GIT_COMMIT)
GO_LDFLAGS_ENTERPRISE = $(GO_LDFLAGS) -X $(GO_VERSION_PATH).feature=enterprise
GO_LDFLAGS_PERSONAL = $(GO_LDFLAGS) -X $(GO_VERSION_PATH).feature=personal

DOCKER_IMAGE ?= codenameuranium/tunnel:$(DOCKER_TAG)-personal
DOCKER_IMAGE_ENTERPRISE ?= codenameuranium/tunnel:$(DOCKER_TAG)
DOCKER_BUILD_ARGS = --progress=plain --platform=linux/amd64


run: run/personal

build: build/personal

docker/build: docker/build/personal

docker/push: docker/push/personal


run/personal: build/personal
	@./tunnel-node

build/personal:
	@echo "+ $@ $(DESCRIPTION) (personal)"
	go build -ldflags="$(GO_LDFLAGS_PERSONAL)" -trimpath -o tunnel-node ./cmd/tunnel/main.go

docker/build/personal:
	@echo "+ $@ $(DOCKER_IMAGE)"
	docker build $(DOCKER_BUILD_ARGS) --tag $(DOCKER_IMAGE) --build-arg TARGET="build/personal" --file ./docker/tunnel/Dockerfile .

docker/push/personal:
	@echo "+ $@ $(DOCKER_IMAGE)"
	@docker push $(DOCKER_IMAGE)

run/enterprise: build/enterprise
	@./tunnel-node

build/enterprise:
	@echo "+ $@ $(DESCRIPTION) (enterprise)"
	go build -ldflags="$(GO_LDFLAGS_ENTERPRISE)" -trimpath -o tunnel-node ./cmd/tunnel/main.go

docker/build/enterprise:
	@echo "+ $@ $(DOCKER_IMAGE_ENTERPRISE)"
	docker build $(DOCKER_BUILD_ARGS) --tag $(DOCKER_IMAGE_ENTERPRISE) --build-arg TARGET="build/enterprise" --file ./docker/tunnel/Dockerfile .

docker/push/enterprise:
	@echo "+ $@ $(DOCKER_IMAGE_ENTERPRISE)"
	@docker push $(DOCKER_IMAGE_ENTERPRISE)


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
	@mv ./proto/github.com/Codename-Uranium/tunnel/proto/*.pb.go ./proto
	@rm -rf ./proto/github.com
