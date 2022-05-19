#!/usr/bin/env bash

kind create cluster --image kindest/node:v1.24.2 --wait 32s --name kusk --config ./cluster.yaml
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.12.1/manifests/namespace.yaml
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.12.1/manifests/metallb.yaml
kubectl apply -f https://kind.sigs.k8s.io/examples/loadbalancer/metallb-configmap.yaml
ANALYTICS_ENABLED=false kusk install --no-api --no-dashboard --name=kusk --namespace=default
kubectl get pods --namespace metallb-system --watch
kind delete cluster --name kusk

cat <<EOF | kind create cluster --image kindest/node:v1.24.2 --wait 256s --name kusk-gateway-dev --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
  - containerPort: 80
    hostPort: 8080
    protocol: TCP
  - containerPort: 443
    hostPort: 4443
    protocol: TCP
  - containerPort: 19000
    hostPort: 19000
    protocol: TCP
# - role: worker
EOF
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.12.1/manifests/namespace.yaml
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.12.1/manifests/metallb.yaml

# $ docker network inspect -f '{{.IPAM.Config}}' kind
# [{172.20.0.0/16  172.20.0.1 map[]} {fc00:f853:ccd:e793::/64   map[]}]
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
      - 172.20.255.200-172.20.255.250
EOF

make install && kind load docker-image kubeshop/kusk-gateway:latest --name kusk-gateway-dev && make deploy && make deploy-envoyfleet

kubectl apply -f https://kind.sigs.k8s.io/examples/loadbalancer/metallb-configmap.yaml
