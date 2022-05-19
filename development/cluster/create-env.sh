#!/usr/bin/env bash
set -e

KUBERNETES_PROVIDER="${KUBERNETES_PROVIDER:=minikube}"

if [[ "${KUBERNETES_PROVIDER}" == "minikube" ]]; then
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
fi

if [[ "${KUBERNETES_PROVIDER}" == "kind" ]]; then
  # sudo sysctl -w kernel.dmesg_restrict=0
  # podman network inspect --format='{{range .}}{{ (index .Subnets 1).Subnet }}{{end}}' kind
  kind create cluster --image kindest/node:v1.24.0 --wait 5m --name kusk --config ./development/cluster/kind/cluster.yaml
  kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.12.1/manifests/namespace.yaml
  kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.12.1/manifests/metallb.yaml
  # kubectl get pods -n metallb-system --watch
  kubectl apply -f https://kind.sigs.k8s.io/examples/loadbalancer/metallb-configmap.yaml
fi

echo "========> installing CRDs"
make install

echo "========> building control-plane docker image and installing into cluster"

SHELL=/bin/bash
make docker-build deploy

kubectl rollout status -w deployment/kusk-gateway-manager -n kusk-system

echo "========> Deploying default Envoy Fleet"
make deploy-envoyfleet
