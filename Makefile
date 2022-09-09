include smoketests/Makefile.variables

.DEFAULT_GOAL := all
MAKEFLAGS += --environment-overrides --warn-undefined-variables # --print-directory --no-builtin-rules --no-builtin-variables

# Add bin to the PATH
export PATH := $(shell pwd)/bin:$(PATH)
BINARIES_DIR := $(shell pwd)/bin
KUSTOMIZE := ${BINARIES_DIR}/kustomize
CONTROLLER_GEN := ${BINARIES_DIR}/controller-gen
PROTOC := ${BINARIES_DIR}/protoc/bin/protoc
PROTOC_GEN_GO := ${BINARIES_DIR}/protoc-gen-go
PROTOC_GEN_GO_GRPC := ${BINARIES_DIR}/protoc-gen-go-grpc
ENVTEST := ${BINARIES_DIR}/setup-envtest
KTUNNEL := ${BINARIES_DIR}/ktunnel
STERN := $(shell go env GOPATH)/bin/stern

VERSION ?= $(shell git describe --tags)

# Image URL to use all building/pushing image targets
IMAGE_TAG ?= dev
MANAGER_IMG ?= kusk-gateway:${IMAGE_TAG}

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION := 1.22

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL 			= /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

LD_FLAGS += -X 'github.com/kubeshop/kusk-gateway/pkg/analytics.TelemetryToken=$(TELEMETRY_TOKEN)'
LD_FLAGS += -X 'github.com/kubeshop/kusk-gateway/pkg/build.Version=$(VERSION)'

# strip DWARF, symbol table and debug info. Expect ~25% binary size decrease
# https://github.com/kubeshop/kusk-gateway/issues/431
LD_FLAGS += -s -w

export DOCKER_BUILDKIT ?= 1

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

ifeq ($(shell uname -s),Linux)
  DOCS_PREVIEWER := xdg-open
else ifeq ($(shell uname -s),Darwin)
  DOCS_PREVIEWER := open
else
  $(error unsupported system: $(shell uname -s))
endif

.PHONY: docs-preview
docs-preview: ## Preview the documentation
	{ cd docs; npm install; ${DOCS_PREVIEWER} http://localhost:3000; npm start & }; wait

.PHONY: tail-logs
tail-logs: $(STERN) ## Tail logs of all containers across all namespaces
	stern --all-namespaces --selector app.kubernetes.io/name

.PHONY: tail-xds
tail-xds: ## Tail logs of kusk-manager
	kubectl logs --follow --namespace kusk-system services/kusk-gateway-xds-service

.PHONY: tail-envoyfleet
tail-envoyfleet: ## Tail logs of envoy
	kubectl logs --follow --namespace default service/default

.PHONY: enable-logging
enable-logging: ## Set some particular logger's level
	kubectl port-forward --namespace default deployments/default 19000:19000 & echo $$! > /tmp/kube-port-forward-logging.pid
	sleep 4
	curl -s -X POST "http://localhost:19000/logging?backtrace=trace"
	curl -s -X POST "http://localhost:19000/logging?envoy_bug=trace"
	curl -s -X POST "http://localhost:19000/logging?assert=trace"
	curl -s -X POST "http://localhost:19000/logging?secret=trace"
	curl -s -X POST "http://localhost:19000/logging?grpc=trace"
	curl -s -X POST "http://localhost:19000/logging?ext_authz=trace"
	curl -s -X POST "http://localhost:19000/logging?filter=trace"
	curl -s -X POST "http://localhost:19000/logging?misc=trace"
	curl -s -X POST "http://localhost:19000/logging?conn_handler=trace"
	@# curl -s -X POST "http://localhost:19000/logging?connection=trace"
	@# curl -s -X POST "http://localhost:19000/logging?http=trace"
	@# curl -s -X POST "http://localhost:19000/logging?http2=trace"
	@# curl -s -X POST "http://localhost:19000/logging?admin=trace"
	@# bash -c "trap 'pkill -F /tmp/kube-port-forward-logging.pid' SIGINT SIGTERM ERR EXIT"
	@echo
	@echo "How to stop port forward to the admin port (19000):"
	@echo "pkill -F /tmp/kube-port-forward-logging.pid"
	@echo

.PHONY: dev-update
dev-update: docker-build update cycle deploy-envoyfleet ## Update cluster with local changes (usually after you have modified the code).

.PHONY: create-env
create-env: ## Spin up a local development cluster with Minikube and install kusk Gateway
	./development/cluster/create-env.sh

.PHONY: deploy-envoyfleet
deploy-envoyfleet: ## Deploy k8s resources for the single Envoy Fleet
	kubectl apply -f config/samples/gateway_v1_envoyfleet.yaml

.PHONY: delete-env
delete-env: ## Destroy the local development Minikube cluster
	minikube delete --profile kgw

.PHONY: manifests
manifests: $(CONTROLLER_GEN) ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=kusk-gateway-manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: $(CONTROLLER_GEN) ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt $(shell go list ./... | grep -v /examples/)

.PHONY: vet
vet: ## Run go vet against code.
	go vet $(shell go list ./... | grep -v /examples/)

.PHONY: test
test: manifests generate fmt vet $(ENVTEST) ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test $(shell go list ./... | grep -v smoketests | grep -v internal/controllers | grep -v api/v1alpha1) -coverprofile cover.out

.PHONY: testing
testing: ## Run the integration tests from development/testing and then delete testing artifacts if succesfull.
	development/testing/runtest.sh all delete

##@ Build

.PHONY: build
build: generate fmt vet ## Build manager binary.
	go build -o bin/manager -ldflags="$(LD_FLAGS)" cmd/manager/main.go

.PHONY: run
run: $(KTUNNEL) install-local generate fmt vet ## Run a controller from your host, proxying it inside the cluster.
	go build -o bin/manager cmd/manager/main.go
	ktunnel expose -n kusk-system kusk-xds-service 18000 & ENABLE_WEBHOOKS=false bin/manager ; fg

.PHONY: docker-build-manager
docker-build-manager: ## Build docker image with the manager.
	eval $$(minikube docker-env --profile kgw); docker build -t ${MANAGER_IMG} -f ./build/manager/Dockerfile .

.PHONY: docker-build
docker-build: docker-build-manager ## Build docker images for all apps

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install-local
install-local: manifests $(KUSTOMIZE) ## Install CRDs and Envoy into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/local | kubectl apply -f -
	kubectl -n kusk-system wait --for condition=established --timeout=60s crd/envoyfleet.gateway.kusk.io
	kubectl -n kusk-system apply -f config/samples/gateway_v1_envoyfleet.yaml

.PHONY: uninstall-local
uninstall-local: manifests $(KUSTOMIZE) ## Uninstall CRDs and Envoy from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/local | kubectl delete -f -

.PHONY: install
install: manifests $(KUSTOMIZE) ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests $(KUSTOMIZE) ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests $(KUSTOMIZE) ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	echo $(MANAGER_IMG)
	$(KUSTOMIZE) build config/default  > tttdeploy.yaml

.PHONY: deploy-debug
deploy-debug: manifests $(KUSTOMIZE) ## Deploy controller with debugger to the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/debug | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl delete -f -

.PHONY: update
update: docker-build-manager deploy cycle ## Runs deploy, docker build and restarts kusk-gateway-manager deployment to pick up the change

.PHONY: cycle
cycle: ## Triggers kusk-gateway-manager deployment rollout restart to pick up the new container image with the same tag
	kubectl rollout restart deployment/kusk-gateway-manager -n kusk-system
	@echo "Triggered deployment/kusk-gateway-manager restart, waiting for it to finish"
	kubectl rollout status deployment/kusk-gateway-manager -n kusk-system --timeout=60s

.PHONY: cycle-envoy
cycle-envoy: ## Triggers all Envoy pods in the cluster to restart
	kubectl rollout restart deployment/kgw-envoy-default -n default
	kubectl rollout restart deployment/kgw-envoy-testing -n testing || echo 'rollout restart failed'

$(KUSTOMIZE): ## Download kustomize locally if necessary.
	GOBIN=${BINARIES_DIR} go install sigs.k8s.io/kustomize/kustomize/v4@v4.5.4

$(CONTROLLER_GEN): ## Download controller-gen locally if necessary.
	GOBIN=${BINARIES_DIR} go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.7.0

$(STERN): ## Download stern (https://github.com/stern/stern) locally if necessary.
	go install github.com/stern/stern@latest
	@# Optionally `source <(stern --completion=zsh)` in your `~/.zshrc`.

# Protoc and friends installation and generation
$(PROTOC):
	$(call install-protoc)

$(PROTOC_GEN_GO):
	@echo "[INFO]: Installing protobuf go generation plugin."
	GOBIN=${BINARIES_DIR} go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.27.1

$(PROTOC_GEN_GO_GRPC):
	@echo "[INFO]: Installing protobuf GRPC go generation plugin."
	GOBIN=${BINARIES_DIR} go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0

# Envtest
$(ENVTEST): ## Download envtest-setup locally if necessary.
	GOBIN=${BINARIES_DIR} go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

$(KTUNNEL):
	GOBIN=${BINARIES_DIR} go install github.com/omrikiei/ktunnel@v1.4.7

# Envtest
ENVTEST = $(shell pwd)/bin/setup-envtest
.PHONY: envtest
envtest: ## Download envtest-setup locally if necessary.
	$(call go-get-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)

.PHONY: tools
tools: $(KTUNNEL) $(ENVTEST) $(PROTOC_GEN_GO_GRPC) $(PROTOC_GEN_GO) $(PROTOC) $(CONTROLLER_GEN) $(KUSTOMIZE) ## Install all tools

define install-protoc
@[ -f "${PROTOC}" ] || { \
set -e ;\
echo "[INFO] Installing protoc compiler to ${BINARIES_DIR}/protoc" ;\
mkdir -pv "${BINARIES_DIR}/" ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
PB_REL="https://github.com/protocolbuffers/protobuf/releases" ;\
VERSION=3.19.4 ;\
if [ "$$(uname)" == "Darwin" ];then FILENAME=protoc-$${VERSION}-osx-x86_64.zip ;fi ;\
if [ "$$(uname)" == "Linux" ];then FILENAME=protoc-$${VERSION}-linux-x86_64.zip;fi ;\
echo "Downloading $${FILENAME} to $${TMP_DIR}" ;\
curl -LO $${PB_REL}/download/v$${VERSION}/$${FILENAME} ; unzip $${FILENAME} -d ${BINARIES_DIR}/protoc ; \
rm -rf $${TMP_DIR} ;\
}
endef

.PHONY: $(smoketests)
$(smoketests):
	$(MAKE) -C smoketests $@

check-all: $(smoketests)
