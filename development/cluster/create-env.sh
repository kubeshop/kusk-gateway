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
make docker-build deploy
kubectl rollout status -w deployment/kusk-gateway-manager -n kusk-system

set +e

echo "========> Deploying default Envoy Fleet"
until make deploy-envoyfleet; do
  # A timing issue sometimes results in the below occuring:
  # Error from server (InternalError): error when creating "config/samples/gateway_v1_envoyfleet.yaml": Internal error occurred: failed calling webhook "menvoyfleet.kb.io": failed to call webhook: Post "https://kusk-gateway-webhooks-service.kusk-system.svc:443/mutate-gateway-kusk-io-v1alpha1-envoyfleet?timeout=10s": dial tcp 10.109.220.117:443: connect: connection refused
  echo 'sleeping for 2 seconds before trying `make deploy-envoyfleet`'
  sleep 2
done
