#!/usr/bin/env bash
# Usage:
# To disable metrics: development/minikube.sh
# To enable metrics: METRICS=1 development/minikube.sh
set -o errexit  # Used to exit upon error, avoiding cascading errors
set -o pipefail # Unveils hidden failures
set -o nounset  # Exposes unset variables

PROFILE="kgw"

function print_separator() {
  local header="${1} "
  local separator_chars=""
  separator_chars=$(printf -- '-%.0s' $(seq 1 $(($(tput cols) - ${#header}))))
  printf "%b" "$(
    tput bold
    tput setaf 2
  )" "${header}" "${separator_chars}" "$(tput sgr0)" "\n"
}

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

print_separator "installing CRDs"
make install

print_separator "building control-plane docker image and installing into cluster"
make docker-build deploy
kubectl rollout status -w deployment/kusk-gateway-manager -n kusk-system

print_separator "Deploying default Envoy Fleet"
until make deploy-envoyfleet; do
  # A timing issue sometimes results in the below occuring:
  # Error from server (InternalError): error when creating "config/samples/gateway_v1_envoyfleet.yaml": Internal error occurred: failed calling webhook "menvoyfleet.kb.io": failed to call webhook: Post "https://kusk-gateway-webhooks-service.kusk-system.svc:443/mutate-gateway-kusk-io-v1alpha1-envoyfleet?timeout=10s": dial tcp 10.109.220.117:443: connect: connection refused
  echo "sleeping for 2 seconds before trying 'make deploy-envoyfleet'"
  sleep 2
done
