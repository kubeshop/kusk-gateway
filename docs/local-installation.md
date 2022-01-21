# Installing Kusk Gateway to the Local Minikube cluster

---
**NOTE**

This is the quick instruction on how to setup the Kusk Gateway and try it out with Minikube on the local machine.

If you're looking for the generic Kubernetes installation instructions, please see [Installation](installation.md).

---

# Table of contents
- [Prerequsities](#prerequsities)
- [Installation](#installation)
- [Uninstallation](#uninstallation)

During the installation we'll create the Minikube Kubernetes cluster, deploy the Kusk Gateway with its Custom Resources Definitions and Envoy Fleet.

After that you'll be able to go through our ToDoMVC example that configures access to API with OpenAPI file and Kusk Gateway.

For the architectural overview of these components please check the [Architecture](arch.md) page.

### Installation requirements

Tools needed for the installation:

- [Minikube](https://minikube.sigs.k8s.io/docs/start/). Make sure you had its *Installation* step finished, you can skip other steps as not needed.
For the MacOS users, the additional configuration step is needed to setup and set as the default for Minikube driver [hyperkit](https://minikube.sigs.k8s.io/docs/drivers/hyperkit/).
- [Helm v3](https://helm.sh/docs/intro/install/)
- [Kubectl](https://kubernetes.io/docs/tasks/tools/)

### Installation

Start the local Minikube cluster

This will switch the default kubectl context to this new cluster kgw.

```sh
# Start cluster. 
minikube start --profile kgw

```

Next we setup all components and Kusk Gateway

Commands below will configure necessary cluster component such as Jetstack Cert-Manager for webhooks configuration and MetalLB that is needed to finish the configuration of the Service with the type **LoadBalancer**.

After that we'll install the Kusk Gateway and the "default" Envoy Fleet (LoadBalancer) in the recommended **kusk-system** namespace with Helm.

We provide 2 Helm [charts](https://github.com/kubeshop/helm-charts):

- **[kusk-gateway](https://github.com/kubeshop/helm-charts/tree/main/charts/kusk-gateway)** chart provides Custom Resources Definitions as well as the Kusk Gateway Manager (Operator) deployment.

- **[kusk-gateway-envoyfleet](https://github.com/kubeshop/helm-charts/tree/main/charts/kusk-gateway-envoyfleet)** chart provides the EnvoyFleet Custom Resource installation, which is used to configure the gateway with KGW Manager.

Container images are hosted on Docker Hub [Kusk-Gateway repository](https://hub.docker.com/r/kubeshop/kusk-gateway).

You can select and copy all of these commands as one block and paste it into the terminal for the speed of installation.

```sh
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
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.6.0/cert-manager.yaml &&\
kubectl wait --for=condition=available --timeout=600s deployment/cert-manager-webhook -n cert-manager

# Install Kubeshop Helm repo and update it
helm repo add kubeshop https://kubeshop.github.io/helm-charts && helm repo update

# Install the Kusk Gateway with CRDs into kusk-system namespace.
# We need to wait for the kusk-gateway-manager deployment to finish the setup for the next step.
helm install kusk-gateway kubeshop/kusk-gateway -n kusk-system --create-namespace &&\
kubectl wait --for=condition=available --timeout=600s deployment/kusk-gateway-manager -n kusk-system

# Install the "default" EnvoyFleet Custom Resource, which will be used by the Kusk Gateway
# to create Envoy Fleet Deployment and Service with the type LoadBalancer
helm install kusk-gateway-envoyfleet kubeshop/kusk-gateway-envoyfleet -n kusk-system

```

This concludes the installation.

It may take a few seconds for the LoadBalancer IP to become available.

Run this to find out the External IP address of EnvoyFleet Load balancer.

```sh
kubectl get svc -l "app=kusk-gateway,component=envoy-svc" --namespace default

```

The output should contain the Service **kusk-envoy-svc-default** with the **External-IP** address field - use this address for your API endpoints querying.

You can now deploy your API or Front applications to this cluster and configure access to them with [Custom Resources](customresources/index.md) or you can check the [ToDoMVC Example](todomvc.md) for the guidelines on how to do this.

In case of the problems please check the [Troubleshooting](troubleshooting.md) section.

### Uninstallation

The following command will delete the Minikube cluster.

```sh
minikube delete --profile kgw
```
