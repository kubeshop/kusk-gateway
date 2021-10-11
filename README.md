# kusk-gateway
Kusk-gateway is the API Gateway, based on Envoy and using OpenAPI specification as the source of configuration

# Steps to setup local development cluster and deploy kusk-gateway operator
- `k3d registry create reg -p 5000`
- `k3d cluster create --registry-use reg cl1`
- add `127.0.0.1 k3d-reg` to /etc/hosts (note the k3d- prefix)
- `kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.5.4/cert-manager.yaml`
- `make docker-build docker-push deploy`