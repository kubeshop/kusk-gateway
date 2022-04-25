# Architecture

Under the hood, Kusk Gateway consists of an Operator (Kusk Gateway Manager) and k8s resources, created and managed by the operator. e.g. APIs, StaticRoutes and EnvoyFleets

Envoy Proxy is deployed by Kusk Gateway Manager as part of [**EnvoyFleet**](customresources/envoyfleet.md) Custom Resource configuration that describes K8s Envoy Proxy deployment and its K8s service.
Usually this service is of type LoadBalancer, thus exposing the service to the world.

One can deploy multiple Envoy Fleets in the scenario when one needs multiple different services on different IP addresses.

Once Envoy Fleet is deployed, Envoy processes connect to the Kusk Gateway Manager for the dynamically updated configuration via GRPC.
Kusk Gatewy Manager accepts CustomResourceDefinitions [**API**](customresources/api.md) and [**StaticRoute**](customresources/staticroute.md) to configure the routing to the deployed in K8s applications services and updates the related Envoy Fleets routing configuration.

![kusk-gateway arch diagram](../img/arch.png)
