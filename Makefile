BINARY = aro-rp-version
COMMIT := $(shell git rev-parse HEAD)
BRANCH := $(shell git symbolic-ref --short -q HEAD || echo HEAD)
DATE := $(shell date -u +%Y%m%d-%H:%M:%S)
VERSION_PKG = github.com/25region/aro-rp-versions/pkg/version
LDFLAGS := "-X ${VERSION_PKG}.Branch=${BRANCH} -X ${VERSION_PKG}.BuildDate=${DATE} \
	-X ${VERSION_PKG}.GitSHA1=${COMMIT}"
TAG?=""

all: build

clean:
	rm -rf $(BINARY) dist/

build:
	CGO_ENABLED=0 go build -o $(BINARY) -ldflags $(LDFLAGS)

vendor:
	go mod vendor

test: lint-all

test-dirty: vendor build
	go mod tidy
	git diff --exit-code

test-release:
	BRANCH=$(BRANCH) COMMIT=$(COMMIT) DATE=$(DATE) VERSION_PKG=$(VERSION_PKG) goreleaser release --snapshot --skip-publish --rm-dist

lint:
	LINT_INPUT="$(shell go list ./...)"; golint -set_exit_status $$LINT_INPUT

golangci-lint:
	golangci-lint run

lint-all: lint golangci-lint

tag:
	git tag -a $(TAG) -m "Release $(TAG)"
	git push origin $(TAG)

# Requires GITHUB_TOKEN environment variable to be set
release:
	BRANCH=$(BRANCH) COMMIT=$(COMMIT) DATE=$(DATE) VERSION_PKG=$(VERSION_PKG) goreleaser release --rm-dist

.PHONY: all clean build vendor image test test-unit test-dirty test-release lint golangci-lint lint-all tag release
