# Add bin to the PATH
export PATH := $(shell pwd)/bin:$(PATH)

# Image URL to use all building/pushing image targets
MANAGER_IMG ?= kusk-gateway:dev
HELPER_IMG ?= kusk-gateway-helper:dev

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.22

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

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

.PHONY: create-env
create-env: ## Spin up a local development cluster with Minikube and install kusk-gateway for debugging
	./development/cluster/create-env.sh

.PHONY: create-cluster
create-cluster: ## Spin up a local minikube cluster with metallb and cert manager installed
	./development/cluster/create-cluster.sh

.PHONY: deploy-envoyfleet
deploy-envoyfleet: ## Deploy k8s resources for the single Envoy Fleet
	kubectl apply -f config/samples/gateway_v1_envoyfleet.yaml

.PHONY: delete-env
delete-env: ## Destroy the local development k3d cluster
	./development/cluster/delete-env.sh

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=kusk-gateway-manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt $(shell go list ./... | grep -v /examples/)

.PHONY: vet
vet: ## Run go vet against code.
	go vet $(shell go list ./... | grep -v /examples/)

.PHONY: test
test: manifests generate fmt vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./... -coverprofile cover.out

.PHONY: testing
testing: ## Run the integration tests from development/testing and then delete testing artifacts if succesfull.
	development/testing/runtest.sh all delete

##@ Build

.PHONY: build
build: generate fmt vet ## Build manager and helper binary.
	go build -o bin/manager cmd/manager/main.go
	go build -o bin/helper cmd/helper/main.go

.PHONY: run
run: install-local generate fmt vet ## Run a controller from your host, proxying it inside the cluster.
	go build -o bin/manager cmd/manager/main.go
	ktunnel expose -n kusk-system kusk-xds-service 18000 & ENABLE_WEBHOOKS=false bin/manager ; fg

.PHONY: docker-build-manager
docker-build-manager: ## Build docker image with the manager.
	@eval $$(minikube docker-env --profile kgw); DOCKER_BUILDKIT=1 docker build -t ${MANAGER_IMG} -f ./build/manager/Dockerfile .

.PHONY: docker-build-helper
docker-build-helper: helper-management-compile ## Build docker image with the helper.
	@eval $$(minikube docker-env --profile kgw); DOCKER_BUILDKIT=1 docker build -t ${HELPER_IMG} -f ./build/helper/Dockerfile .

.PHONY: docker-build
docker-build: docker-build-manager docker-build-helper ## Build docker images for all apps

.PHONY: docker-build-manager-debug
docker-build-manager-debug: ## Build docker image with the manager and debugger.
	@eval $$(SHELL=/bin/bash minikube docker-env --profile kgw) ;DOCKER_BUILDKIT=1 docker build -t "${MANAGER_IMG}-debug" -f ./build/manager/Dockerfile-debug .

.PHONY: docker-build-debug-helper
docker-build-debug-helper: helper-management-compile ## Build docker image with the helper and debugger.
	@eval $$(SHELL=/bin/bash minikube docker-env --profile kgw) ;DOCKER_BUILDKIT=1 docker build -t "${HELPER_IMG}-debug" -f ./build/helper/Dockerfile-debug .

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install-local
install-local: manifests kustomize ## Install CRDs and Envoy into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/local | kubectl apply -f -
	kubectl -n kusk-system wait --for condition=established --timeout=60s crd/envoyfleet.gateway.kusk.io
	kubectl -n kusk-system apply -f config/samples/gateway_v1_envoyfleet.yaml

.PHONY: uninstall-local
uninstall-local: manifests kustomize ## Uninstall CRDs and Envoy from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/local | kubectl delete -f -

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: deploy-debug
deploy-debug: manifests kustomize ## Deploy controller with debugger to the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/debug | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl delete -f -

.PHONY: update
update: docker-build-manager deploy cycle ## Runs deploy, docker build and restarts kusk-gateway-manager deployment to pick up the change

.PHONY: update-helper
update-helper: docker-build-helper cycle-envoy

.PHONY: update-debug
update-debug: docker-build-manager-debug deploy-debug cycle ## Runs Debug configuration deploy, docker build and restarts kusk-gateway-manager deployment to pick up the change

.PHONY: cycle
cycle: ## Triggers kusk-gateway-manager deployment rollout restart to pick up the new container image with the same tag
	kubectl rollout restart deployment/kusk-gateway-manager -n kusk-system
	@echo "Triggered deployment/kusk-gateway-manager restart, waiting for it to finish"
	kubectl rollout status deployment/kusk-gateway-manager -n kusk-system --timeout=30s

.PHONY: cycle-envoy
cycle-envoy: ## Triggers all Envoy pods in the cluster to restart
	kubectl rollout restart deployment/kgw-envoy-default -n default
	-kubectl rollout restart deployment/kgw-envoy-testing -n testing

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
.PHONY: controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.7.0)

KUSTOMIZE = $(shell pwd)/bin/kustomize
.PHONY: kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

# Protoc and friends installation and generation
PROTOC := $(shell which protoc)
PROTOC_GEN_GO := $(shell pwd)/bin/protoc-gen-go

# If protoc isn't on the path, set it to a target that's never up to date, so
# the install command always runs.
ifeq ($(PROTOC),)
	PROTOC = must-rebuild
endif
# Figure out which machine we're running on.
UNAME := $(shell uname)

$(PROTOC):
	# Run the right installation command for the operating system.
	ifeq ($(UNAME), Darwin)
		brew install protobuf
	endif
	ifeq ($(UNAME), Linux)
		@echo "[INFO]: Installing protobuf-compiler, input your password for sudo."
		sudo apt-get install protobuf-compiler -y
	endif
$(PROTOC_GEN_GO):
	@echo "[INFO]: Installing protobuf go generation plugin."
	$(call go-get-tool,$(PROTOC_GEN_GO),google.golang.org/protobuf/cmd/protoc-gen-go@v1.27.1)


	# $(PROTOC) --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative $(notdir $(PROTOFILE))
	echo $(dirname $@)

.PHONY: helper-management-compile
helper-management-compile: $(PROTOC) $(PROTOC_GEN_GO) # Compile protoc files for helper/management
	cd "internal/helper/management";$(PROTOC) --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative *.proto

# Envtest
ENVTEST = $(shell pwd)/bin/setup-envtest
.PHONY: envtest
envtest: ## Download envtest-setup locally if necessary.
	$(call go-get-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef
