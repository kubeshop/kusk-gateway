# Installing Kusk Gateway

## Helm installation

We provide [Helm v3](https://helm.sh/) [charts](https://github.com/kubeshop/helm-charts) for the Kusk Gateway installation.

There are 2 charts to install:

* **[kusk-gateway](https://github.com/kubeshop/helm-charts/tree/main/charts/kusk-gateway)** chart provides Custom Resources Definitions as well as Kusk Gateway Manager (Operator) deployment.

* **[kusk-gateway-envoyfleet](https://github.com/kubeshop/helm-charts/tree/main/charts/kusk-gateway-envoyfleet)** chart provides the EnvoyFleet Custom Resource installation, which is used to configure the gateway with KGW Manager.

Container images are hosted on Docker Hub [Kusk-Gateway repository](https://hub.docker.com/r/kubeshop/kusk-gateway).

### Prerequsities

* [Helm v3](https://helm.sh/) and [Kubectl](https://kubernetes.io/docs/tasks/tools/) installed

* Kubernetes cluster administration rights are required - we install CRDs, service account with ClusterRoles and RoleBindings.

* We heavily depend on [jetstack cert-manager](https://github.com/jetstack/cert-manager) for webhooks TLS configuration. If it is not installed in your cluster, then please install it with the official instructions [here](https://cert-manager.io/docs/installation/).

* If you try to install the chart to your local machine cluster (k3d or minikube), you may need to install and configure [MetalLB](https://metallb.universe.tf/) to handle LoadBalancer type services,
otherwise EnvoyFleet service ExternalIP address will be in Pending state forever. See installing [Locally with Minikube and Helm](#locally-with-minikube-and-helm) 

### Installation

The commands below will install Kusk Gateway and the "default" Envoy Fleet (LoadBalancer) in the recommended **kusk-system** namespace.

```sh
helm repo add kubeshop https://kubeshop.github.io/helm-charts
helm repo update
helm install kusk-gateway kubeshop/kusk-gateway -n kusk-system --create-namespace
kubectl wait --for=condition=available --timeout=600s deployment/kusk-gateway-manager  -n kusk-system
helm install kusk-gateway-default-envoyfleet kubeshop/kusk-gateway-envoyfleet -n kusk-system
```

You can now deploy your Gateway Custom resources, see also Examples section below.

### Uninstallation

```sh
helm delete kusk-gateway-default-envoyfleet kusk-gateway -n kusk-system && kubectl delete namespace kusk-system
```

## Locally with Minikube and Helm

### Prerequisities

Installed:

* [Minikube](https://minikube.sigs.k8s.io/docs/)
* [Helm v3](https://helm.sh/)
* [Kubectl](https://kubernetes.io/docs/tasks/tools/)

### Minikube cluster setup and installation with Helm

Run the following set of commands to quickly setup Minikube-based K8s cluster and setup Kusk Gateway with Helm.

```sh
minikube start --profile kgw

# determine load balancer ingress range
cidr_base_addr=$(minikube ip --profile kgw)
ingress_first_addr=$(echo "$cidr_base_addr" | awk -F'.' '{print $1,$2,$3,2}' OFS='.')
ingress_last_addr=$(echo "$cidr_base_addr" | awk -F'.' '{print $1,$2,$3,255}' OFS='.')
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

# Install Jetstack Cert Manager and wait for it to be ready
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.6.0/cert-manager.yaml
kubectl wait --for=condition=available --timeout=600s deployment/cert-manager-webhook -n cert-manager

# Finally install KGW
helm repo add kubeshop https://kubeshop.github.io/helm-charts
helm repo update
helm install kusk-gateway kubeshop/kusk-gateway -n kusk-system --create-namespace
kubectl wait --for=condition=available --timeout=600s deployment/kusk-gateway-manager  -n kusk-system
helm install kusk-gateway-default-envoyfleet kubeshop/kusk-gateway-envoyfleet -n kusk-system
```

Run this to find out External IP address of EnvoyFleet Load balancer that was setup by MetalLB.

NOTE: It may take a seconds for the LoadBalancer IP to be available.

```sh
kubectl get svc -l "app=kusk-gateway,component=envoy-svc,fleet=default" --namespace kusk-system
```

You can now use found EXTERNAL-IP in your URLs.

If you want to try deploying the example, please run the following:

```sh
kubectl apply -f https://raw.githubusercontent.com/kubeshop/kusk-gateway/main/examples/httpbin/manifest.yaml && kubectl rollout status -w deployment/httpbin

kubectl apply -f https://raw.githubusercontent.com/kubeshop/kusk-gateway/main/examples/httpbin/httpbin_v1_api.yaml

kubectl apply -f https://raw.githubusercontent.com/kubeshop/kusk-gateway/main/examples/httpbin/httpbin_v1_staticroute.yaml

# Wait few seconds for KGW to finish the configuration.
# This should return
curl -v http://<YOUR_EXTERNAL_IP>:8080/get
```

To uninstall everything - just delete that cluster.

```sh
minikube delete --profile kgw
```

## Locally with manifests files and Docker build

This the approach used for the development.

### Prerequisites

Make sure you have the following installed and on your PATH:

- `jq`
- `kubectl`
- `docker`
- `minikube`

Run:

- `make create-cluster` # creates and configures the minikube cluster
- `make install` # install the required CRDs
- `kubectl apply -f ./config/samples/gateway_v1_envoyfleet.yaml -n kusk-system` to install the envoy fleet
- `eval $(minikube docker-env --profile "kgw")` # so built docker images are available to Minikube
- `make docker-build deploy` # build and deploy the kusk gateway image
- `kubectl rollout status -w deployment/kusk-controller-manager -n kusk-system`

Once Kusk Gateway is installed and running, you can try and apply your own OpenAPI specs, or apply one of our examples below.

### Example

This example will deploy httpbin application and configure Kusk Gateway to serve it.

```sh
kubectl apply -f examples/httpbin && kubectl rollout status -w deployment/httpbin

external_ip=$(kubectl -n kusk-system get svc kusk-envoy --template="{{range .status.loadBalancer.ingress}}{{.ip}}{{end}}")
curl -v http://$external_ip:8080/get
```
