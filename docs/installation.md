# Installing Kusk Gateway

---
**NOTE**

This documents describes the installation of the Kusk Gateway and its load balancing component Envoy Fleet to the generic Kubernetes cluster.

If you're looking for the quick way to try Kusk Gateway in a locally setup Kubernetes cluster, please see [Local Installation with Minikube](local-installation.md).

---

# Table of contents
- [Prerequsities](#prerequsities)
  - [Cluster requirements](#cluster-requirements)
  - [Install requirements](#install-requirements)
- [Installation](#installation)
- [Uninstallation](#uninstallation)

During the setup we'll deploy Kusk Gateway Custom Resources Definitions, Kusk Gateway Manager and Envoy Fleet with Helm.

For the architectural overview of the components please check the [Architecture](arch.md) page.

## Prerequsities

### Cluster requirements

- Kubernetes v1.16+

- Kubernetes cluster administration rights are required - we install CRDs, service account with ClusterRoles and RoleBindings.

- If you don't have Jetstack Cert-Manager installed in your cluster, then please follow the official [instructions](https://cert-manager.io/docs/installation/) to setup it. We use Cert-Manager for the webhooks configuration.

- If you have the managed cluster (GCP, EKS, etc) then you can skip to the next section.
If you have the baremetal or locally setup cluster, then you should have the controller that manages load balancing setup when a Service with the type **LoadBalancer** is setup. Otherwise when the Manager creates the Envoy Fleet Service, it will have stuck ExternalIP address in a Pending state forever. [MetalLB](https://metallb.universe.tf/installation/) provides such functionality, so we advise to setup it if you haven't already.

### Installation requirements

Tools needed for the installation:

- [Helm v3](https://helm.sh/docs/intro/install/)
- [Kubectl](https://kubernetes.io/docs/tasks/tools/)

### Installation

We provide 2 Helm [charts](https://github.com/kubeshop/helm-charts):

- **[kusk-gateway](https://github.com/kubeshop/helm-charts/tree/main/charts/kusk-gateway)** chart provides Custom Resources Definitions as well as the Kusk Gateway Manager (Operator) deployment.

- **[kusk-gateway-envoyfleet](https://github.com/kubeshop/helm-charts/tree/main/charts/kusk-gateway-envoyfleet)** chart provides the EnvoyFleet Custom Resource installation, which is used to configure the gateway with KGW Manager.

Container images are hosted on Docker Hub [Kusk-Gateway repository](https://hub.docker.com/r/kubeshop/kusk-gateway).

The commands below will install Kusk Gateway and the "default" Envoy Fleet (LoadBalancer) in the recommended **kusk-system** namespace.

```sh
# Install Kubeshop Helm repo and update it
helm repo add kubeshop https://kubeshop.github.io/helm-charts
helm repo update

# Install the Kusk Gateway with CRDs into kusk-system namespace.
helm install kusk-gateway kubeshop/kusk-gateway -n kusk-system --create-namespace

# We need to wait for the kusk-gateway-manager deployment to finish the setup for the next step.
kubectl wait --for=condition=available --timeout=600s deployment/kusk-gateway-manager  -n kusk-system

# Install the "default" EnvoyFleet Custom Resource, which will be used by the Kusk Gateway
# to create Envoy Fleet Deployment and Service with the type LoadBalancer
helm install kusk-gateway-default-envoyfleet kubeshop/kusk-gateway-envoyfleet -n kusk-system
```

This concludes the installation

It may take a few seconds for the LoadBalancer IP to become available.

Run this to find out the External IP address of EnvoyFleet Load balancer.

```sh
kubectl get svc -l "app=kusk-gateway,component=envoy-svc,fleet=default" --namespace kusk-system
```

The output should contain the Service **kusk-envoy-svc-default** with the **External-IP** address field - use this address for your API endpoints querying.

You can now deploy your API or Front applications to this cluster and configure access to them with [Custom Resources](customresources/index.md) or you can check the [ToDoMVC Example](todomvc.md) for the guidelines on how to do this.

In case of the problems please check the [Troubleshooting](troubleshooting.md) section.

### Uninstallation

The following command will uninstall Kusk Gateway with CRDs and the **default** Envoy Fleet with their namespace.

```sh
# Delete releases
helm delete kusk-gateway-default-envoyfleet kusk-gateway -n kusk-system

# Delete namespace too
kubectl delete namespace kusk-system
```
