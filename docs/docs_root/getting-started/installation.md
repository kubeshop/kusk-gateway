# Installing Kusk Gateway

## **Prerequisites**

- Kubernetes v1.16+
- Kubernetes Cluster Administration rights are required - we 
  install [CustomResourceDefinitions](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#customresourcedefinitions) 
  and a ServiceAccount with ClusterRoles and RoleBindings.

## **Installation requirements**

Tools needed for the installation:

- [helm](https://helm.sh/docs/intro/install/) command-line tool
- [kubectl](https://kubernetes.io/docs/tasks/tools/) command-line tool

## **Installing Kusk Gateway**
### **1. Install Kusk CLI** 

You can find other installation methods (like Homebrew) [here](../cli/overview.md).

```sh
bash < <(curl -sSLf https://raw.githubusercontent.com/kubeshop/kusk/main/scripts/install.sh)
```

### **2. Install Kusk Gateway**

Use the Kusk CLIs [install command](../cli/install-cmd.md) to install Kusk Gateway in your cluster. 

```sh
kusk install
```

### **3. Access the Dashboard**

Kusk Gateway includes a [browser-based dashboard](../dashboard/overview.md) for inspection and management of your deployed APIs.
Use the following commands to open it in your local browser after the above installation finishes.

```shell
kubectl port-forward -n kusk-system svc/kusk-gateway-private-envoy-fleet 8080:80
open http://localhost:8080
```

## **Get the Gateway's External IP**

If you want to access the APIs or StaticRoutes managed by Kusk Gateway, get the External IP address of the 
Load Balancer by running the command below. Note that it may take a few seconds for the LoadBalancer IP to become available.

```sh
kubectl get svc -l "app.kubernetes.io/component=envoy-svc" --namespace kusk-system
```

The output should contain the [Envoy Fleet](../customresources/envoyfleet) Service, which is the entry point of your API gateway, with the **External-IP** address field - use this address for your API endpoints querying. Note that it might take a while for the External IP to be created.

!!! note non-important "External IP might not be available for some cluster setups".

    If you are running a **local setup**, you can access the API endpoint with: 
    
    `kubectl port-forward service/kusk-gateway-envoy-fleet 8088:80 -n kusk-system`

    If you are running a **bare metal cluster**, consider installing [MetalLB](https://metallb.universe.tf) which creates External IP for LoadBalancer Service type in Kubernetes.

If there are any issues, please check the [Troubleshooting](../guides/troubleshooting.md) section.
