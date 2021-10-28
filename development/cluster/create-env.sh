#!/usr/bin/env bash

set -e

echo "creating registry k3d-reg:5000. add 127.0.0.1 k3d-reg to /etc/hosts (note the k3d- prefix)"
k3d registry create reg -p 5000

echo "creating cluster..."
k3d cluster create local-k8s --servers 1 --agents 1 --registry-use reg --k3s-arg "--disable=traefik@server:0" -p "8080:8080@loadbalancer" --wait

# determine load balancer ingress range
cidr_block=$(docker network inspect k3d-local-k8s | jq '.[0].IPAM.Config[0].Subnet' | tr -d '"')
cidr_base_addr=${cidr_block%???}
ingress_first_addr=$(echo "$cidr_base_addr" | awk -F'.' '{print $1,$2,255,0}' OFS='.')
ingress_last_addr=$(echo "$cidr_base_addr" | awk -F'.' '{print $1,$2,255,255}' OFS='.')
ingress_range=$ingress_first_addr-$ingress_last_addr

# deploy metallb
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.10.2/manifests/namespace.yaml
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.10.2/manifests/metallb.yaml

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

echo "installing cert manager"
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.5.4/cert-manager.yaml

echo "installing CRDs"
make install

echo "installing "
make docker-build docker-push deploy

echo "installing httpbin"
kubectl apply -f examples/httpbin/manifest.yaml
kubectl apply -f examples/httpbin/httpbin_v1_api.yaml

