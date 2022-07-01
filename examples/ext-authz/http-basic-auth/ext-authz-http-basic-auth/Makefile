# MIT License
#
# Copyright (c) 2022 Kubeshop
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

.DEFAULT_GOAL := all
MAKEFLAGS += --environment-overrides --warn-undefined-variables # --print-directory --no-builtin-rules --no-builtin-variables
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

IMAGE_NAME := kubeshop/ext-authz-http-basic-auth

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
	go build -tags 'osusergo netgo' -ldflags '-extldflags "-static" -s -w' -a -v -o ./ext-authz-http-basic-auth ./...

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
	$(CONTAINER_ENGINE) run --rm --interactive --tty ${IMAGE_NAME}:latest

.PHONY: image-push
image-push: image-build
	@# $(CONTAINER_ENGINE) push docker.io/${IMAGE_NAME}:latest
	@# $(CONTAINER_ENGINE) push docker.io/${IMAGE_NAME}:v0.1.0
	$(CONTAINER_ENGINE) push ttl.sh/${IMAGE_NAME}:24h
