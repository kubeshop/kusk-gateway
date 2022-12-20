# Static Route

Static Routes define the entrypoint to static applications.
This overcomes the limitation of OpenAPI specification that does not allow to define the entrypoint to a static application.
It is useful to set up the routing to a non-API application, e.g. static pages or images or to route to legacy, possibly external to the cluster, APIs.

Once the resource manifest is deployed, Kusk Gateway Manager will use it to configure routing for the Envoy Fleet.

## **Limitations**

Static Routes assume entrypoints to your upstream hosts or services live at `/`
For this reason, each `StaticRoute` must have its own dedicated `EnvoyFleet`

Currently, the resource **status** field is not updated by the manager when the reconciliation of the configuration finishes.

## **Configuration Structure Description**

The main elements of the configuration are in the **spec** field.

Below is the YAML structure of the configuration. Please read further for a full explanation.

```yaml
apiVersion: gateway.kusk.io/v1alpha1
kind: StaticRoute
metadata:
  name: staticroute-sample
spec:
  # Envoy Fleet (its name and namespace) to deploy the configuration to, here - deployed EnvoyFleet with the name "default" in the namespace "default".
  # Optional, if not specified - single (default) fleet autodetection will be performed in the cluster.
  fleet:
    name: default
    namespace: default
  hosts: [<string>, <string>, ...]
  auth:
    # ouath2 | jwt | cloudentity | custom
   ...
  upstream:
    # host | service | rewrite
   service:
     name: my-service
     namespace: my-namespace
     port: 80
...
```

## **Envoy Fleet**

The spec.**fleet** optional field specifies to which Envoy Fleet (Envoy Proxy instances with the exposing K8s Service) this configuration applies.
fleet.**name** and fleet.**namespace** reference the deployed Envoy Fleet Custom Resource name and namespace.

You can deploy a Static Route configuration in any namespace with any name and it will be applied to the specific Envoy Fleet.
If this option is missing, the autodetection will be performed to find the single fleet deployed in the Kubernetes cluster Fleet, which is then considered as the default Fleet.
The deployed Static Route custom resource will be changed to map to that fleet accordingly.
If there are multiple fleets deployed, spec.**fleet** is required to specify which in the manifest.

## **Request Matching**

We match the incoming request by HOST header, path and HTTP method.

The following fields specify matching:

**hosts** - Defines the list of HOST headers to which the current configuration applies. This will create the Envoy's VirtualHost with the same name and domain matching. Wildcards are possible, e.g. "*" means "any host".
Prefix and suffix wildcards are supported, but not both (i.e. ```example.*, *example.com```, but not ```*example*```).

**upstream** - Defines the upstream host or service to which the request will be forwarded.

## **Example**

```yaml
apiVersion: gateway.kusk.io/v1alpha1
kind: StaticRoute
metadata:
  name: sample
spec:
  # should work with localhost, example.org, any host
  hosts: [ "localhost", "*"]
  auth:
    oauth2:
      issuer: https://auth.example.com
      client_id: client_id
      client_secret: client_secret
      scopes:
        - openid
        - email
        - profile
      callback_path: /callback
      callback_url: https://example.com/callback
      redirect_url: https://example.com
      token_path: /oauth2/token
      user_info_path: /userinfo
      user_info_url: https://auth.example.com/userinfo
      user_info_mapping:
        email: email
        name: name
        picture: picture
  upstream:
    host: "httpbin.org"
```
