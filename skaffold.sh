#!/usr/bin/env bash
# Usage:
# To disable metrics: development/minikube.sh
# To enable metrics: METRICS=1 development/minikube.sh
set -o errexit # Used to exit upon error, avoiding cascading errors
set -o nounset # Exposes unset variables
set -x

PROFILE="${PROFILE:-kgw}"

install_and_configure_skaffold() {
  ARCH="$([ $(uname -m) = "aarch64" ] && echo "arm64" || echo "amd64")"
  sudo curl -L --output /usr/local/bin/skaffold "https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-${ARCH}"
  sudo chmod +x /usr/local/bin/skaffold
  echo
  skaffold version
  echo
  skaffold config set --global local-cluster true
  echo
  mkdir -pv /tmp/skaffold || echo '`/tmp/skaffold` already exist - skipping create'
}
skaffold version || install_and_configure_skaffold

run_kustomize() {
  kustomize build config/crd >/tmp/skaffold/config-crd.yaml
  # For debugging support changing this value, otherwise we get this error:
  # `message: 'container has runAsNonRoot and image will run as root (pod: "kusk-gateway-manager-67cdb6b9d6-6scdk_kusk-system(dfd51e59-eac6-483d-8b58-52be68f824dc)",`
  kustomize build config/default | sed -E 's/runAsNonRoot: true/runAsNonRoot: false/g' >/tmp/skaffold/config-default.yaml
}
run_kustomize

load_balancer_minikube() {
  # # Determine load balancer ingress range
  # CIDR_BASE_ADDR="$(minikube ip --profile "${PROFILE}")"
  # INGRESS_FIRST_ADDR="$(echo "${CIDR_BASE_ADDR}" | awk -F'.' '{print $1,$2,$3,2}' OFS='.')"
  # INGRESS_LAST_ADDR="$(echo "${CIDR_BASE_ADDR}" | awk -F'.' '{print $1,$2,$3,255}' OFS='.')"
  # INGRESS_RANGE="${INGRESS_FIRST_ADDR}-${INGRESS_LAST_ADDR}"
  # # configure metallb ingress address range
  # echo "${CONFIG_MAP_METALLB}" >/tmp/skaffold/config-map-metallb.yaml
  echo
}
load_balancer_minikube

load_balancer_kind() {
  mkdir -pv /tmp/skaffold/metallb

  KIND_NET_CIDR=$(docker network inspect kind -f '{{(index .IPAM.Config 0).Subnet}}')
  METALLB_IP_START=$(echo ${KIND_NET_CIDR} | sed "s@0.0/16@255.200@")
  METALLB_IP_END=$(echo ${KIND_NET_CIDR} | sed "s@0.0/16@255.250@")
  METALLB_IP_RANGE="${METALLB_IP_START}-${METALLB_IP_END}"

  CONFIG_MAP_METALLB="apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: example
  namespace: metallb-system
spec:
  addresses:
  - ${METALLB_IP_RANGE}
---
apiVersion: metallb.io/v1beta1
kind: L2Advertisement
metadata:
  name: empty
  namespace: metallb-system"

  echo "${CONFIG_MAP_METALLB}" >/tmp/skaffold/metallb/metallb-config.yaml
  curl --silent -L --output /tmp/skaffold/metallb/metallb-native.yaml https://raw.githubusercontent.com/metallb/metallb/v0.13.5/config/manifests/metallb-native.yaml
}
load_balancer_kind

skaffold "${@}"
