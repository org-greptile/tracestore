# Version number
VERSION=$(shell ./tools/image-tag | cut -d, -f 1)

GIT_REVISION := $(shell git rev-parse --short HEAD)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

GOPATH := $(shell go env GOPATH)
GORELEASER := $(GOPATH)/bin/goreleaser

# Build Images
DOCKER_PROTOBUF_IMAGE ?= otel/build-protobuf:0.14.0
LOGSTORE_BUILD_IMAGE ?= acme/logstore-build-image:0.21.0
DOCS_IMAGE ?= acme/docs-base:latest

# More exclusions can be added similar with: -not -path './testbed/*'
ALL_SRC := $(shell find . -name '*.go' \
								-not -path './vendor*/*' \
								-not -path './integration/*' \
								-not -path './cmd/tracestore-serverless/*' \
                                -type f | sort)

# All source code and documents. Used in spell check.
ALL_DOC := $(shell find . \( -name "*.md" -o -name "*.yaml" \) \
                                -type f | sort)

# ALL_PKGS is used with 'go cover'
ALL_PKGS := $(shell go list $(sort $(dir $(ALL_SRC))))

GO_OPT= -mod vendor -ldflags "-X main.Branch=$(GIT_BRANCH) -X main.Revision=$(GIT_REVISION) -X main.Version=$(VERSION)"
ifeq ($(BUILD_DEBUG), 1)
	GO_OPT+= -gcflags="all=-N -l"
endif

GOTEST_OPT?= -race -timeout 30m -count=1
GOTEST_OPT_WITH_COVERAGE = $(GOTEST_OPT) -cover
GOTEST=go test
LINT=golangci-lint

UNAME := $(shell uname -s)
ifeq ($(UNAME), Darwin)
    SED_OPTS := ''
endif

FILES_TO_FMT=$(shell find . -type d \( -path ./vendor -o -path ./opentelemetry-proto -o -path ./vendor-fix \) -prune -o -name '*.go' -not -name "*.pb.go" -not -name '*.y.go' -print)
FILES_TO_JSONNETFMT=$(shell find ./operations/jsonnet ./operations/tracestore-mixin -type f \( -name '*.libsonnet' -o -name '*.jsonnet' \) -not -path "*/vendor/*" -print)

### Build

.PHONY: tracestore
tracestore:
	GO111MODULE=on CGO_ENABLED=0 go build $(GO_OPT) -o ./bin/$(GOOS)/tracestore-$(GOARCH) $(BUILD_INFO) ./cmd/tracestore

.PHONY: tracestore-query
tracestore-query:
	GO111MODULE=on CGO_ENABLED=0 go build $(GO_OPT) -o ./bin/$(GOOS)/tracestore-query-$(GOARCH) $(BUILD_INFO) ./cmd/tracestore-query

.PHONY: tracestore-cli
tracestore-cli:
	GO111MODULE=on CGO_ENABLED=0 go build $(GO_OPT) -o ./bin/$(GOOS)/tracestore-cli-$(GOARCH) $(BUILD_INFO) ./cmd/tracestore-cli

.PHONY: tracestore-vulture
tracestore-vulture:
	GO111MODULE=on CGO_ENABLED=0 go build $(GO_OPT) -o ./bin/$(GOOS)/tracestore-vulture-$(GOARCH) $(BUILD_INFO) ./cmd/tracestore-vulture

.PHONY: exe
exe:
	GOOS=linux $(MAKE) $(COMPONENT)

.PHONY: exe-debug
exe-debug:
	BUILD_DEBUG=1 GOOS=linux $(MAKE) $(COMPONENT)

### Testin' and Lintin'

.PHONY: test
test:
	$(GOTEST) $(GOTEST_OPT) $(ALL_PKGS)

.PHONY: benchmark
benchmark:
	$(GOTEST) -bench=. -run=notests $(ALL_PKGS)

.PHONY: test-with-cover
test-with-cover: test-serverless
	$(GOTEST) $(GOTEST_OPT_WITH_COVERAGE) $(ALL_PKGS)

# runs e2e tests in the top level integration/e2e directory
.PHONY: test-e2e
test-e2e: docker-tracestore
	$(GOTEST) -v $(GOTEST_OPT) ./integration/e2e

# runs only serverless e2e tests
.PHONY: test-e2e-serverless
test-e2e-serverless: docker-tracestore docker-serverless
	$(GOTEST) -v $(GOTEST_OPT) ./integration/e2e/serverless

# test-all/bench use a docker image so build it first to make sure we're up to date
.PHONY: test-all
test-all: test-with-cover test-e2e test-e2e-serverless

.PHONY: test-bench
test-bench: docker-tracestore
	$(GOTEST) -v $(GOTEST_OPT) ./integration/bench

.PHONY: fmt check-fmt
fmt:
	@gofmt -s -w $(FILES_TO_FMT)
	@goimports -w $(FILES_TO_FMT)

check-fmt: fmt
	@git diff --exit-code -- $(FILES_TO_FMT)

.PHONY: jsonnetfmt check-jsonnetfmt
jsonnetfmt:
	@jsonnetfmt -i $(FILES_TO_JSONNETFMT)

check-jsonnetfmt: jsonnetfmt
	@git diff --exit-code -- $(FILES_TO_JSONNETFMT)

.PHONY: lint
lint:
	$(LINT) run

### Docker Images

.PHONY: docker-component # Not intended to be used directly
docker-component: check-component exe
	docker build -t acme/$(COMPONENT) --build-arg=TARGETARCH=$(GOARCH) -f ./cmd/$(COMPONENT)/Dockerfile .
	docker tag acme/$(COMPONENT) $(COMPONENT)

.PHONY: docker-component-debug
docker-component-debug: check-component exe-debug
	docker build -t acme/$(COMPONENT)-debug --build-arg=TARGETARCH=$(GOARCH) -f ./cmd/$(COMPONENT)/Dockerfile_debug .
	docker tag acme/$(COMPONENT)-debug $(COMPONENT)-debug

.PHONY: docker-tracestore
docker-tracestore:
	COMPONENT=tracestore $(MAKE) docker-component

docker-tracestore-debug:
	COMPONENT=tracestore $(MAKE) docker-component-debug

.PHONY: docker-cli
docker-tracestore-cli:
	COMPONENT=tracestore-cli $(MAKE) docker-component

.PHONY: docker-tracestore-query
docker-tracestore-query:
	COMPONENT=tracestore-query $(MAKE) docker-component

.PHONY: docker-tracestore-vulture
docker-tracestore-vulture:
	COMPONENT=tracestore-vulture $(MAKE) docker-component

.PHONY: docker-images
docker-images: docker-tracestore docker-tracestore-query docker-tracestore-vulture

.PHONY: check-component
check-component:
ifndef COMPONENT
	$(error COMPONENT variable was not defined)
endif

# #########
# Gen Proto
# #########
PROTOC = docker run --rm -u ${shell id -u} -v${PWD}:${PWD} -w${PWD} ${DOCKER_PROTOBUF_IMAGE} --proto_path=${PWD}
PROTO_INTERMEDIATE_DIR = pkg/.patched-proto
PROTO_INCLUDES = -I$(PROTO_INTERMEDIATE_DIR)
PROTO_GEN = $(PROTOC) $(PROTO_INCLUDES) --gogofaster_out=plugins=grpc,paths=source_relative:$(2) $(1)
PROTO_GEN_WITH_VENDOR = $(PROTOC) $(PROTO_INCLUDES) -Ivendor -Ivendor/github.com/gogo/protobuf --gogofaster_out=plugins=grpc,paths=source_relative:$(2) $(1)

.PHONY: gen-proto
gen-proto:
	@echo --
	@echo -- Deleting existing
	@echo --
	rm -rf opentelemetry-proto
	rm -rf $(PROTO_INTERMEDIATE_DIR)
	find pkg/tracestorepb -name *.pb.go | xargs -L 1 -I rm
	# Here we avoid removing our tracestore.proto and our frontend.proto due to reliance on the gogoproto bits.
	find pkg/tracestorepb -name *.proto | grep -v tracestore.proto | grep -v frontend.proto | xargs -L 1 -I rm

	@echo --
	@echo -- Copying to $(PROTO_INTERMEDIATE_DIR)
	@echo --
	git submodule update --init
	mkdir -p $(PROTO_INTERMEDIATE_DIR)
	cp -R opentelemetry-proto/opentelemetry/proto/* $(PROTO_INTERMEDIATE_DIR)

	@echo --
	@echo -- Editing proto
	@echo --

	@# Update package and types from opentelemetry.proto.* -> tracestorepb.*
	@# giving final types like "tracestorepb.common.v1.InstrumentationLibrary" which
	@# will not conflict with other usages of opentelemetry proto in downstream apps.
	find $(PROTO_INTERMEDIATE_DIR) -name "*.proto" | xargs -L 1 sed -i $(SED_OPTS) 's+ opentelemetry.proto+ tracestorepb+g'

	@# Update go_package
	find $(PROTO_INTERMEDIATE_DIR) -name "*.proto" | xargs -L 1 sed -i $(SED_OPTS) 's+go.opentelemetry.io/proto/otlp+example.com/acme/tracestore/pkg/tracestorepb+g'

	@# Update import paths
	find $(PROTO_INTERMEDIATE_DIR) -name "*.proto" | xargs -L 1 sed -i $(SED_OPTS) 's+import "opentelemetry/proto/+import "+g'

	@echo --
	@echo -- Gen proto --
	@echo --
	$(call PROTO_GEN,$(PROTO_INTERMEDIATE_DIR)/common/v1/common.proto,./pkg/tracestorepb/)
	$(call PROTO_GEN,$(PROTO_INTERMEDIATE_DIR)/resource/v1/resource.proto,./pkg/tracestorepb/)
	$(call PROTO_GEN,$(PROTO_INTERMEDIATE_DIR)/trace/v1/trace.proto,./pkg/tracestorepb/)
	$(call PROTO_GEN,pkg/tracestorepb/tracestore.proto,./)
	$(call PROTO_GEN_WITH_VENDOR,modules/frontend/v1/frontendv1pb/frontend.proto,./)

	rm -rf $(PROTO_INTERMEDIATE_DIR)

# ##############
# Gen Traceql
# ##############
.PHONY: gen-traceql
gen-traceql:
	docker run --rm -v${PWD}:/src/logstore ${LOGSTORE_BUILD_IMAGE} gen-traceql-local

.PHONY: gen-traceql-local
gen-traceql-local:
	goyacc -o pkg/traceql/expr.y.go pkg/traceql/expr.y && rm y.output

### Check vendored and generated files are up to date
.PHONY: vendor-check
vendor-check: gen-proto update-mod gen-traceql
	git diff --exit-code -- **/go.sum **/go.mod vendor/ pkg/tracestorepb/ pkg/traceql/

### Tidy dependencies for tracestore and tracestore-serverless modules
.PHONY: update-mod
update-mod:
	go mod vendor
	go mod tidy -e
	$(MAKE) -C cmd/tracestore-serverless update-mod


### Release (intended to be used in the .github/workflows/release.yml)
$(GORELEASER):
	go install github.com/goreleaser/goreleaser@latest

.PHONY: release
release: $(GORELEASER)
	$(GORELEASER) release --rm-dist

.PHONY: release-snapshot
release-snapshot: $(GORELEASER)
	$(GORELEASER) release --skip-validate --rm-dist --snapshot

### Docs
.PHONY: docs
docs:
	docker pull ${DOCS_IMAGE}
	docker run -v ${PWD}/docs/sources/tracestore:/hugo/content/docs/tracestore/latest:z -p 3002:3002 --rm $(DOCS_IMAGE) /bin/bash -c 'mkdir -p content/docs/acme/latest/ && touch content/docs/acme/latest/menu.yaml && make server'

.PHONY: docs-test
docs-test:
	docker pull ${DOCS_IMAGE}
	docker run -v ${PWD}/docs/sources/tracestore:/hugo/content/docs/tracestore/latest:z -p 3002:3002 --rm $(DOCS_IMAGE) /bin/bash -c 'mkdir -p content/docs/acme/latest/ && touch content/docs/acme/latest/menu.yaml && make prod'

### jsonnet
.PHONY: jsonnet jsonnet-check
jsonnet:
	$(MAKE) -C operations/jsonnet-compiled/util gen

jsonnet-check:
	$(MAKE) -C operations/jsonnet-compiled/util check


### serverless
.PHONY: docker-serverless test-serverless
docker-serverless:
	$(MAKE) -C cmd/tracestore-serverless build-docker

test-serverless:
	$(MAKE) -C cmd/tracestore-serverless test

### tracestore-mixin
.PHONY: tracestore-mixin tracestore-mixin-check
tracestore-mixin:
	$(MAKE) -C operations/tracestore-mixin all

tracestore-mixin-check:
	$(MAKE) -C operations/tracestore-mixin check

### drone
.PHONY: drone drone-jsonnet drone-signature
# this requires the drone-cli https://docs.drone.io/cli/install/
drone:
	# piggyback on Logstore's build image, this image contains a newer version of drone-cli than is
	# released currently (1.4.0). The newer version of drone-clie keeps drone.yml human-readable.
	# This will run 'make drone-jsonnet' from within the container
	docker run -e DRONE_SERVER -e DRONE_TOKEN --rm -v $(shell pwd):/src/logstore ${LOGSTORE_BUILD_IMAGE} drone-jsonnet drone-signature

	drone lint .drone/drone.yml --trusted

drone-jsonnet:
	drone jsonnet --stream --format --source .drone/drone.jsonnet --target .drone/drone.yml

drone-signature:
ifndef DRONE_TOKEN
	$(error DRONE_TOKEN is not set, visit https://drone.acme.net/account)
endif
	DRONE_SERVER=https://drone.acme.net drone sign --save acme/tracestore .drone/drone.yml
