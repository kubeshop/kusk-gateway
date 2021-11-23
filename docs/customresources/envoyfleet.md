# Envoy Fleet

This resource defines EnvoyFleet, which is the implementation of gateway in Kubernetes based on Envoy Proxy.

Once resource manifest is deployed to Kubernets, it is used by Kusk Gateway Manager to setup K8s Envoy Proxy **Deployment** and **Service** with the type *LoadBalancer*.

Envoy Proxy is configured to connect for the configuration to [XDS](https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol) service of KGW Manager.
In its initial state there is no configuration, you have to deploy API or StaticRoute resource to setup the routing.

If the Custom Resource is uninstalled, Manager deletes created K8s resouces.

Currently supported parameters:

* spec.**size** , which is the number of Envoy Proxy pods in K8s deployment, defaults to 1 if not specified.
* metadata.**name** is the EnvoyFleet ID. Manager will supply the configuration for this specific ID.

**Alpha Limitations**:

* the support for multiple IDs configuration is a work in progress, right now used the single ID **default**.

* currently resource **status** field is not updated by manager when the configuration process finishes.

```yaml EnvoyFleet.yaml
apiVersion: gateway.kusk.io/v1alpha1
kind: EnvoyFleet
metadata:
  name: default
spec:
  size: 1
```
