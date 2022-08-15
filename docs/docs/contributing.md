# How to Develop Kusk Gateway

Kusk Gateway code is managed with the help of [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) that provides code scaffolding and generation of K8s Custom Resource Definitions files.

Internally, Kusk Gateway uses the [go-control-plane](https://github.com/envoyproxy/go-control-plane) package to configure Envoy with its xDS protocol.

## **Code Structure to Get Started with Development**
Below is the (reduced) output of the `tree` command on this repository.

These are the directories and packages that we suggest you need to know about to get started with Kusk Gateway development.

Obviously, there are more, so feel free to investigate the others for yourself to get an idea about how they fit into the overall architecture.

```
kusk-gateway
├── api
│   └── v1alpha1 # Our Custom CRD types in Go
├── build # Dockerfiles for the Manager and Agent processes
│   ├── agent
│   └── manager
├── cmd # Entry point for the Manager and Agent processes
│   ├── agent
│   └── manager
├── examples # Various example applications to test against
│   ├── httpbin
│   ├── todomvc
│   │   └── spec
│   └── websocket
├── internal
│   ├── agent # Business Logic for the agent process which handles mocking API resposes
│   │   ├── httpserver
│   │   ├── management
│   │   └── mocking
│   ├── controllers # Custom Kubernetes controllers that handle CRD events
│   ├── envoy # Envoy configuration, go-control-plane manager set up, and the types we manage when building the config
│   │   ├── config
│   │   ├── manager
│   │   └── types
│   ├── k8sutils # useful helper functions for interacting with Kubernetes
│   ├── validation # Service proxy that handles request validation before forwarding the request onto the destination service
│   └── webhooks # Create certs for the webhooks
└── pkg
    ├── analytics # Code for sending analytics data to telemetry provider
    ├── options # Options structs that contain the fields that the user will configure with the x-kusk extension in their OpenAPI definition
    └── spec # Code for loading, parsing and validating the OpenAPI definition and the extensions.
```

## **Set Up Development Environment**
You can install Kusk Gateway into any cluster offering. So you can BYOC (Bring Your Own Cluster).

### **Launch Kusk Gateway in Minikube**
We have a useful Make command to set up a complete environment for development in Minikube.

To set up the environment, run the following command:

```
make create-env
```

This will do the following:

- Start minikube with the profile name "kusk" and enable the Metallb add on. Metallb will expose the Envoy Fleet services "locally" without needing to port-forward to them.   
- Install our CustomResourceDefinitions.   
- Build the docker images, cache them for faster rebuilds and deploy them with Kustomize.   
- Deploy an Envoy Fleet.   

### **Launch Kusk Gateway in Your Cluster**
If you opt to use a cluster offering that is not Minikube, you can use the following commands to launch Kusk Gateway in your cluster:

```
# Install the CustomResourceDefinitions
make install

# Build and deploy the containers using Kustomize
make docker-images-cache docker-build deploy

# wait for rollout to complete
kubectl rollout status -w deployment/kusk-gateway-manager -n kusk-system

# Deploy EnvoyFleet
make deploy-envoyfleet
```

## **Redeploy After Code Changes**
The simplest way to ensure the recompliation of everything required is to run one of the following commands:

```
# generate will recompile any controller implementations
# manifest will regenerate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects that may need updating
# install will install the most recent versions of the CRDs
# docker-build will rebuild the docker images
# deploy will deploy the most recent versions of the containers
# cycle will cycle the Kubernetes deployments to make sure they are running the most recent built versions

make generate manifests install docker-build deploy cycle
```
