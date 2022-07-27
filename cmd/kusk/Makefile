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

LD_FLAGS += -w -s
LD_FLAGS += -X 'github.com/kubeshop/kusk-gateway/pkg/build.Version=${VERSION}'
LD_FLAGS += -X 'github.com/kubeshop/kusk-gateway/pkg/analytics.TelemetryToken=$(TELEMETRY_TOKEN)'

# Determine if we should use:
# 1. docker and docker-compose, or
# 2. podman and podman-compose
CONTAINER_ENGINE ?= $(shell docker version >/dev/null 2>&1 && which docker)
ifeq ($(CONTAINER_ENGINE),)
	CONTAINER_ENGINE = $(shell podman version >/dev/null 2>&1 && which podman)
endif

.PHONY: all
all: install-tools pre-commit

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
	go test -count=1 -v -race ./...

.PHONY: build
build:
	go build -v -o ./kusk -ldflags="${LD_FLAGS}" ./main.go

# `build-goreleaser` is just for local testing.
.PHONY: build-goreleaser
build-goreleaser:
	VERSION="${VERSION}" TELEMETRY_TOKEN="${TELEMETRY_TOKEN}" GA_SECRET="${GA_SECRET}" goreleaser release --rm-dist --skip-publish --skip-validate
