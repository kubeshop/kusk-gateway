# Architecture

Under the hood, Kusk Gateway consists of an Operator (Kusk Gateway Manager) and k8s resources, 
created and managed by the operator, e.g. APIs, Static Routes and Envoy Fleets.

Envoy Proxy is deployed by Kusk Gateway Manager as part of [**Envoy Fleet**](../customresources/envoyfleet.md) 
Custom Resource configuration that describes K8s Envoy Proxy deployment and its K8s service.
Usually this service is of type LoadBalancer, thus exposing the service to the world.

Multiple Envoy Fleets can be deployed in the scenario when multiple heterogenous services exist on different IP addresses.

Once Envoy Fleet is deployed, Envoy processes connect to the Kusk Gateway Manager for the dynamically updated configuration via GRPC.
Kusk Gateway Manager accepts CustomResourceDefinitions [**API**](../customresources/api.md) 
and [**Static Route**](../customresources/staticroute.md) to configure the routing deployed in the K8s applications services 
and updates the related Envoy Fleets routing configuration.

![kusk-gateway arch diagram](../img/arch.png)
