#!/usr/bin/env bash
set -o errexit  # Used to exit upon error, avoiding cascading errors
set -o pipefail # Unveils hidden failures
set -o nounset  # Exposes unset variables
set -x

SCRIPT_DIR="$(dirname "$0")"

# # Kubernetes v1.25.3 image
# kind create cluster \
#   --image kindest/node:v1.25.2@sha256:f52781bc0d7a19fb6c405c2af83abfeb311f130707a0e219175677e366cc45d1 \
#   --wait 256s \
#   --name "${CLUSTER_NAME:-kgw}"

kubectl apply -f "${SCRIPT_DIR}/metallb-native.yaml"

kubectl wait --namespace metallb-system \
  --for=condition=ready pod \
  --selector=app=metallb \
  --timeout=256s

# Alternative is to use: `docker network inspect -f '{{.IPAM.Config}}' kind`
CIDR_START="$(docker network inspect kind -f '{{(index .IPAM.Config 0).Subnet}}' | sed "s@0.0/16@255.200@")"
CIDR_END="$(docker network inspect kind -f '{{(index .IPAM.Config 0).Subnet}}' | sed "s@0.0/16@255.250@")"

cat <<EOF | kubectl apply -f -
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: example
  namespace: metallb-system
spec:
  addresses:
    - ${CIDR_START}-${CIDR_END}
---
apiVersion: metallb.io/v1beta1
kind: L2Advertisement
metadata:
  name: empty
  namespace: metallb-system
EOF
