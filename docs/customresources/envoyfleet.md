# Envoy Fleet

This resource defines the EnvoyFleet, which is the implementation of the gateway in Kubernetes based on Envoy Proxy.

Once the resource manifest is deployed to Kubernetes, it is used by Kusk Gateway Manager to setup K8s Envoy Proxy **Deployment** and **Service** with the type *LoadBalancer*.

Envoy Proxy is configured to connect to the [XDS](https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol) service of the KGW Manager to retrieve the configuration.
In its initial state there is no configuration, you have to deploy API or StaticRoute resource to setup the routing.

If the Custom Resource is uninstalled, the Manager deletes the created K8s resources.

Currently supported parameters:

* spec.**size** , which is the number of Envoy Proxy pods in the K8s deployment, defaults to 1 if not specified.
* metadata.**name** is the EnvoyFleet ID. The Manager will supply the configuration for this specific ID, matching the API / StaticRoute by fleet name.

**Alpha Limitations**:

* the support for multiple IDs configuration is a work in progress, right now the static ID **default** is used.

* currently resource **status** field is not updated by the manager when the reconsiliation of the configuration finishes.

```yaml EnvoyFleet.yaml
apiVersion: gateway.kusk.io/v1alpha1
kind: EnvoyFleet
metadata:
  name: default
spec:
  size: 1
```
