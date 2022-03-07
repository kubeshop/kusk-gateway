#!/usr/bin/env bash

set -e

echo "========> creating cluster..."
minikube start --profile kgw --addons=metallb

# determine load balancer ingress range
cidr_base_addr=$(minikube ip --profile kgw)
ingress_first_addr=$(echo "$cidr_base_addr" | awk -F'.' '{print $1,$2,$3,2}' OFS='.')
ingress_last_addr=$(echo "$cidr_base_addr" | awk -F'.' '{print $1,$2,$3,255}' OFS='.')
ingress_range=$ingress_first_addr-$ingress_last_addr

# configure metallb ingress address range
cat <<EOF | kubectl apply -f -
apiVersion: v1
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
      - $ingress_range
EOF

echo "========> installing CRDs"
make install

echo "========> building control-plane docker image and installing into cluster"

SHELL=/bin/bash
make docker-images-cache docker-build deploy

kubectl rollout status -w deployment/kusk-gateway-manager -n kusk-system

echo "========> Deploying default Envoy Fleet"
make deploy-envoyfleet
