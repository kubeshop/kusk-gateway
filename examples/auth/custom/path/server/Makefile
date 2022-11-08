.DEFAULT_GOAL := all
MAKEFLAGS += --environment-overrides --warn-undefined-variables
SHELL := /usr/bin/env bash -eufo pipefail -c

export DOCKER_BUILDKIT ?= 1
export TERM ?= xterm-256
export CGO_ENABLED := 0

# Determine if we should use:
# 1. `docker` and `docker-compose`, or
# 2. `podman` and `podman-compose`
CONTAINER_ENGINE ?= $(shell docker version >/dev/null 2>&1 && which docker)
ifeq ($(CONTAINER_ENGINE),)
	CONTAINER_ENGINE = $(shell podman version >/dev/null 2>&1 && which podman)
endif

#IMAGE_NAME := kubeshop/auth-custom-path
#IMAGE_NAME := kubeshop-io-auth-custom-path
IMAGE_NAME := banaioltd/kubeshop-io-auth-custom-path

.PHONY: all
all: pre-commit

.PHONY: pre-commit
pre-commit: format lint test build

.PHONY: format
format:
	go mod tidy -v
	gofmt -e -s -w .

.PHONY: lint
lint:
	go vet ./...

.PHONY: test
test:
	CGO_ENABLED=1 go test -count=1 -v -race ./...

.PHONY: build
build:
	go build -ldflags '-extldflags "-static" -s -w' -v -o ./main ./...

.PHONY: image-build
image-build:
	$(CONTAINER_ENGINE) build \
		--tag ${IMAGE_NAME}:v0.1.0 \
		--tag ${IMAGE_NAME}:latest \
		--tag ttl.sh/${IMAGE_NAME}:24h \
		--file ./Dockerfile \
		.

.PHONY: image-run
image-run: image-build
	$(CONTAINER_ENGINE) run --rm --publish 8080:8080 --interactive --tty ${IMAGE_NAME}:latest

.PHONY: image-push
image-push: image-build
	@# $(CONTAINER_ENGINE) push docker.io/${IMAGE_NAME}:latest
	@# $(CONTAINER_ENGINE) push docker.io/${IMAGE_NAME}:v0.1.0
	@# $(CONTAINER_ENGINE) push ttl.sh/${IMAGE_NAME}:24h
	$(CONTAINER_ENGINE) push docker.io/${IMAGE_NAME}:v0.1.0
	$(CONTAINER_ENGINE) push docker.io/${IMAGE_NAME}:v0.1.0
