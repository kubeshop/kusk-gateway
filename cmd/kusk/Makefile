.DEFAULT_GOAL := all
MAKEFLAGS += --environment-overrides --warn-undefined-variables #--print-directory --no-builtin-rules --no-builtin-variables

SHELL := /bin/bash
ifneq ($(shell uname),Darwin)
	SHELL += -O globstar -O extglob
endif

.SHELLFLAGS := -eu -o pipefail -c

export TERM ?= xterm-256
export PATH := $(shell go env GOPATH)/bin:${PATH}

VERSION ?= $(shell git describe --tags)

LD_FLAGS += -X 'github.com/kubeshop/kusk-gateway/pkg/build.Version=${VERSION}'
LD_FLAGS += -X 'github.com/kubeshop/kusk-gateway/pkg/analytics.TelemetryToken=${TELEMETRY_TOKEN}'
# # MBana: Don't strip symbol table and debug info for now.
# LD_FLAGS += -w -s

IMAGE_TAG ?= ${VERSION}
MANAGER_IMG ?= kubeshop/kusk-gateway:${IMAGE_TAG}

# Determine if we should use: docker or podman.
CONTAINER_ENGINE ?= $(shell docker version >/dev/null 2>&1 && which docker)
ifeq ($(CONTAINER_ENGINE),)
	CONTAINER_ENGINE = $(shell podman version >/dev/null 2>&1 && which podman)
endif

.PHONY: all
all: install-tools install-deps pre-commit

.PHONY: install-tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.46.2
	go install github.com/goreleaser/goreleaser@v1.9.1
	go install mvdan.cc/sh/v3/cmd/shfmt@v3.5.0

.PHONY: pre-commit
pre-commit: format lint test build

.PHONY: format
format:
	go mod tidy -v
	@echo
	gofmt -e -s -w .
	@echo
	shfmt -w ./**/*.sh
	@echo
	-$(CONTAINER_ENGINE) run --rm --interactive --tty --workdir /kusk --volume "$(shell pwd)":/kusk:z docker.io/koalaman/shellcheck-alpine:stable sh -c "shellcheck ./**/*.sh"

.PHONY: lint
lint:
	-golangci-lint run --config .golangci.yml ./...

.PHONY: test
test:
	go test -count=1 ./...

.PHONY: install-deps
install-deps:
	go install github.com/go-bindata/go-bindata/v3/go-bindata@v3.1.3
	go install sigs.k8s.io/kustomize/kustomize/v4@v4.5.2

manifests: install-deps
	cd ../../config/default && kustomize edit set image kusk-gateway=${MANAGER_IMG}
	go generate ./...

.PHONY: build
build: manifests
	go build -o ./kusk -ldflags="${LD_FLAGS}" ./main.go

# `build-goreleaser` is just for local testing.
.PHONY: build-goreleaser
build-goreleaser:
	VERSION="${VERSION}" TELEMETRY_TOKEN="${TELEMETRY_TOKEN}" goreleaser release --rm-dist --skip-publish --skip-validate
