#!/usr/bin/env bash
# Usage:
# To disable metrics: development/minikube.sh
# To enable metrics: METRICS=1 development/minikube.sh
set -o errexit  # Used to exit upon error, avoiding cascading errors
set -o pipefail # Unveils hidden failures
set -o nounset  # Exposes unset variables

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

PROFILE="${PROFILE:-kgw}"

kustomize build config/crd >/tmp/skaffold/config-crd.yaml
# For debugging support changing this value, otherwise we get this error:
# `message: 'container has runAsNonRoot and image will run as root (pod: "kusk-gateway-manager-67cdb6b9d6-6scdk_kusk-system(dfd51e59-eac6-483d-8b58-52be68f824dc)",`
kustomize build config/default | sed -E 's/runAsNonRoot: true/runAsNonRoot: false/g' >/tmp/skaffold/config-default.yaml

# Determine load balancer ingress range
CIDR_BASE_ADDR="$(minikube ip --profile "${PROFILE}")"
INGRESS_FIRST_ADDR="$(echo "${CIDR_BASE_ADDR}" | awk -F'.' '{print $1,$2,$3,2}' OFS='.')"
INGRESS_LAST_ADDR="$(echo "${CIDR_BASE_ADDR}" | awk -F'.' '{print $1,$2,$3,255}' OFS='.')"
INGRESS_RANGE="${INGRESS_FIRST_ADDR}-${INGRESS_LAST_ADDR}"

CONFIG_MAP_METALLB="apiVersion: v1
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
      - ${INGRESS_RANGE}"

# configure metallb ingress address range
echo "${CONFIG_MAP_METALLB}" >/tmp/skaffold/config-map-metallb.yaml

skaffold "${@}"
