# Installing Kusk Gateway

## Helm
TODO

## Locally
If you want to run kusk-gateway locally, you can do this easily using Minikube

# Prerequisites
make sure you have the following installed and on your PATH:
- `jq`
- `kubectl`
- `docker`
- `minikube`

Run:
- `make create-cluster` # creates and configures the minikube cluster
- `make install` # install the required CRDs
- `eval $(minikube docker-env --profile "kgw")` # so built docker images are available to Minikube
- `make docker-build deploy` # build and deploy the kusk gateway image
- `kubectl rollout status -w deployment/kusk-controller-manager -n kusk-system`

Once Kusk Gateway is installed and running, you can try and apply your own OpenAPI specs, see Usage below or you can apply one of our examples

# Example
```
kubectl apply -f examples/httpbin && kubectl rollout status -w deployment/httpbin

external_ip=$(kubectl -n kusk-system get svc kusk-envoy --template="{{range .status.loadBalancer.ingress}}{{.ip}}{{end}}")
curl -v http://$external_ip:8080/get
```
