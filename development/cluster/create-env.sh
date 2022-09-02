#!/usr/bin/env bash
# Usage:
# To disable metrics: development/minikube.sh
# To enable metrics: METRICS=1 development/minikube.sh
set -o errexit  # Used to exit upon error, avoiding cascading errors
set -o pipefail # Unveils hidden failures
set -o nounset  # Exposes unset variables

PROFILE="kgw"
SLEEP="4"

# print_separator
# Prints "-" until it reaches end of terminal width.
print_separator() {
  local header="${1} "
  local separator_chars=""
  separator_chars=$(printf -- '-%.0s' $(seq 1 $(($(tput cols) - ${#header}))))
  printf "%b" "$(
    tput bold
    tput setaf 2
  )" "${header}" "${separator_chars}" "$(tput sgr0)" "\n"
}

# create_cluster
# Creates a cluster with either metrics enabled or not and a LoadBalancer, in this case, metallb.
create_cluster() {
  print_separator "creating cluster"
  # If `METRICS` environmental variable has been set, then enable metrics on the cluster.
  if [[ ! -z "${METRICS:=}" ]]; then
    (
      set -x
      minikube start --profile "${PROFILE}" --addons metallb --addons metrics-server
      kubectl -n kube-system rollout status deployment metrics-server
    )
  else
    (
      set -x
      minikube start --profile "${PROFILE}" --addons metallb
    )
  fi

  # determine load balancer ingress range
  CIDR_BASE_ADDR="$(minikube ip --profile "${PROFILE}")"
  INGRESS_FIRST_ADDR="$(echo "${CIDR_BASE_ADDR}" | awk -F'.' '{print $1,$2,$3,2}' OFS='.')"
  INGRESS_LAST_ADDR="$(echo "${CIDR_BASE_ADDR}" | awk -F'.' '{print $1,$2,$3,255}' OFS='.')"
  INGRESS_RANGE="${INGRESS_FIRST_ADDR}-${INGRESS_LAST_ADDR}"

  CONFIG_MAP="apiVersion: v1
kind: ConfigMap
metadata:
  namespace: metallb-system
  name: config
data:
  config: |
    address-pools:
    - name: default
      protocol: layer2
      addresses:
      - $INGRESS_RANGE"

  # configure metallb ingress address range
  echo "${CONFIG_MAP}" | kubectl apply -f -
}

# install_crds
# Installs the CRDs.
install_crds() {
  print_separator "installing CRDs"
  make install
}

# deploy
# build `kusk-gateway` image
build() {
  print_separator "building control-plane image"
  make docker-build
}

# deploy
# deploy `kusk-gateway` image into cluster.
deploy() {
  print_separator "installing control-plane image into cluster"
  make deploy
  kubectl rollout status deployment/kusk-gateway-manager --namespace kusk-system --watch --timeout=64s

  print_separator "Deploying default Envoy Fleet"
  until make deploy-envoyfleet; do
    # A timing issue sometimes results in the below occuring:
    # Error from server (InternalError): error when creating "config/samples/gateway_v1_envoyfleet.yaml": Internal error occurred: failed calling webhook "menvoyfleet.kb.io": failed to call webhook: Post "https://kusk-gateway-webhooks-service.kusk-system.svc:443/mutate-gateway-kusk-io-v1alpha1-envoyfleet?timeout=10s": dial tcp 10.109.220.117:443: connect: connection refused
    echo "sleeping for 2 seconds before trying 'make deploy-envoyfleet' again ..."
    sleep "${SLEEP}"
  done

  sleep "${SLEEP}"
  kubectl rollout status deployment/default --namespace default --watch --timeout=64s
}

create_cluster
install_crds
build
deploy