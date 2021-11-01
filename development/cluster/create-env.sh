#!/usr/bin/env bash

set -e

echo "========> creating cluster..."
minikube start --profile kgw

# determine load balancer ingress range
cidr_block=$(docker network inspect kgw | jq '.[0].IPAM.Config[0].Subnet' | tr -d '"')
cidr_base_addr=${cidr_block%???}
ingress_first_addr=$(echo "$cidr_base_addr" | awk -F'.' '{print $1,$2,255,0}' OFS='.')
ingress_last_addr=$(echo "$cidr_base_addr" | awk -F'.' '{print $1,$2,255,255}' OFS='.')
ingress_range=$ingress_first_addr-$ingress_last_addr

# deploy metallb
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.11.0/manifests/namespace.yaml
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.11.0/manifests/metallb.yaml

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

echo "========> installing cert manager"
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.6.0/cert-manager.yaml

echo "========> waiting for cert manager to become ready"
kubectl wait --for=condition=available --timeout=600s deployment/cert-manager-webhook -n cert-manager

echo "========> installing CRDs"
make install


echo "========> building control-plane docker image and installing into cluster"
$(minikube docker-env --profile kgw)
make docker-build deploy
$(minikube docker-env --profile kgw -u)

kubectl rollout status -w deployment/kusk-controller-manager -n kusk-system
