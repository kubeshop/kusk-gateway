# Installing Kusk Gateway

## Prerequisites

- Kubernetes v1.16+

- Kubernetes Cluster Administration rights are required - we install [CustomResouseDefinitions](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#customresourcedefinitions), service account with ClusterRoles and RoleBindings.

## Installation requirements

Tools needed for the installation:

- Installed [helm](https://helm.sh/docs/intro/install/) command-line
- Installed [kubectl](https://kubernetes.io/docs/tasks/tools/) command-line tool

## 1. Install Kusk CLI

```sh
brew install kusk
```

## 2. Install Kusk Gateway

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

## 3. Get the Gateway's External IP

To get the External IP address of the Load Balancer run the command below command. Note that it may take a few seconds for the LoadBalancer IP to become available.

```sh
kubectl get svc -l "app.kubernetes.io/component=envoy-svc" --namespace kusk-system
```

The output should contain the [Envoy Fleet](https://kubeshop.github.io/kusk-gateway/customresources/envoyfleet) Service, which is the entry point of your API gateway, with the **External-IP** address field - use this address for your API endpoints querying. Note that it might take a while for the External IP to be created.

!!! note non-important "External IP might not be available for some cluster setups"

    If you are running a local setup with **Minikube**, you can access the API endpoint with `minikube service kusk-gateway-envoyfleet -n kusk-system`

    If you are running a **bare metal cluster**, consider installing [MetalLB](https://metallb.universe.tf) which creates External IP for LoadBalancer Service type in Kubernetes.

In case of the problems please check the [Troubleshooting](troubleshooting.md) section.
