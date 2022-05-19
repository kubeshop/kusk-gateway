#!/usr/bin/env bash

kind create cluster --image kindest/node:v1.24.0 --wait 5m --name kusk --config ./cluster.yaml
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.12.1/manifests/namespace.yaml
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.12.1/manifests/metallb.yaml
kubectl apply -f https://kind.sigs.k8s.io/examples/loadbalancer/metallb-configmap.yaml
ANALYTICS_ENABLED=false kusk install --no-api --no-dashboard --name=kusk --namespace=default
kubectl get pods --namespace metallb-system --watch
kind delete cluster --name kusk
