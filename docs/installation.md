# Installing Kusk Gateway

---
**NOTE**

This documents describes the installation of the Kusk Gateway and its load balancing component Envoy Fleet to the generic Kubernetes cluster.

If you're looking for the quick way to try Kusk Gateway in a locally setup Kubernetes cluster, please see [Local Installation with Minikube](local-installation.md).

---

# Table of contents
- [Installing Kusk Gateway](#installing-kusk-gateway)
- [Table of contents](#table-of-contents)
  - [Prerequsities](#prerequsities)
    - [Cluster requirements](#cluster-requirements)
    - [Installation requirements](#installation-requirements)
    - [Installation](#installation)
    - [Uninstallation](#uninstallation)

During the setup we'll deploy Kusk Gateway Custom Resources Definitions, Kusk Gateway Manager and Envoy Fleet with Helm.

For the architectural overview of the components please check the [Architecture](arch.md) page.

## Prerequsities

### Cluster requirements

- Kubernetes v1.16+

- Kubernetes cluster administration rights are required - we install CRDs, service account with ClusterRoles and RoleBindings.

- If you have the managed cluster (GCP, EKS, etc) then you can skip to the next section.
If you have the baremetal or locally setup cluster, then you should have the controller that manages load balancing setup when a Service with the type **LoadBalancer** is setup. Otherwise when the Manager creates the Envoy Fleet Service, it will have stuck ExternalIP address in a Pending state forever. [MetalLB](https://metallb.universe.tf/installation/) provides such functionality, so we advise to setup it if you haven't already.

### Installation requirements

Tools needed for the installation:

- [Helm v3](https://helm.sh/docs/intro/install/)
- [Kubectl](https://kubernetes.io/docs/tasks/tools/)

### Installation

We provide 2 Helm [charts](https://github.com/kubeshop/helm-charts):

- **[kusk-gateway](https://github.com/kubeshop/helm-charts/tree/main/charts/kusk-gateway)** chart provides Custom Resources Definitions as well as the Kusk Gateway Manager (Operator) deployment. You can install only one such chart into a cluster.

- **[kusk-gateway-envoyfleet](https://github.com/kubeshop/helm-charts/tree/main/charts/kusk-gateway-envoyfleet)** chart provides the EnvoyFleet Custom Resource installation, which is used to configure the gateway with KGW Manager. You can install multiple releases of fleets into a cluster.

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

# Install EnvoyFleet into kusk-system namespace. It will be used by the Kusk Gateway
# to create Envoy Fleet Deployment and Service with the type LoadBalancer.
helm install kusk-gateway-envoyfleet kubeshop/kusk-gateway-envoyfleet -n kusk-system
```

This concludes the installation.

This Envoy fleet will be used for all further deployed API and StaticRoutes unless you want to create multiple sites with different IP addresses.
In such a case install another release with a different name and/or namespace. Beware that you'll have to specify the fleet to bind to in your API/StaticRoutes after that.

To get the External IP address of the Load Balancer run the command below command. Note that it may take a few seconds for the LoadBalancer IP to become available.

```sh
kubectl get svc -l "app.kubernetes.io/part-of=kusk-gateway,app.kubernetes.io/component=envoy-svc" --namespace kusk-system
```

The output should contain the Envoy Fleet Service with the **External-IP** address field - use this address for your API endpoints querying.

You can now deploy your API or Front applications to this cluster and configure access to them with [Custom Resources](customresources/index.md) or you can check the [ToDoMVC Example](todomvc.md) for the guidelines on how to do this.

In case of the problems please check the [Troubleshooting](troubleshooting.md) section.

### Uninstallation

The following command will uninstall Kusk Gateway with CRDs and the Envoy Fleet with their namespace.

```sh
# Delete releases

helm delete kusk-gateway-envoyfleet -n kusk-system
helm delete kusk-gateway -n kusk-system

# Delete namespace too
kubectl delete namespace kusk-system
```
