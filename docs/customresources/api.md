# API

This resource uses OpenAPI file with x-kusk annotations as the source of truth to configure routing.
Refer to [extention](../extension.md) for the further information on how to add routing information to OpenAPI file.

The required field of API resource is spec.**spec** where changed OpenAPI file is supplied as a string.

Once the resource manifest is deployed, Kusk Gateway Manager will use it to configure routing for Envoy Fleet.
Multiple resources can exist in different namespaces, all of them will be evaluated and the configuration merged on any action with the separate resource.
Deployment of the resource that has conflicting with the existing resources route (path + HTTP method) will be declined.

**Alpha Limitations**:

* currently resource **status** field is not updated by manager when the configuration process finishes.

*Example*

```yaml
apiVersion: gateway.kusk.io/v1alpha1
kind: API
metadata:
  name: api-sample
spec:
  # OpenAPI file with x-kusk annotation here
  spec: |
    openapi: 3.0.2
    servers:
      - url: /api/v3
    info:
      description: Some description
      version: 1.0.0
      title: the best API in the world
    # top level x-kusk extension to configure routes
    x-kusk:
      disabled: false
      hosts: [ "*" ]
      cors:
        origins:
        - "*"
        methods:
        - POST
        - GET
        - OPTIONS
        headers:
        - Content-Type
        credentials: true
        expose_headers:
        - X-Custom-Header1
        max_age: 86200
      upstream:
        service:
          name: oldapi
          namespace: default
          port: 80
      path:
        prefix: /api
        # Strips prefix when forwarding to upstream
        rewrite:
          pattern: "^/api"
          substitution: ""
    paths:
      /pet:
        x-kusk:
          disabled: true
        post:
          x-kusk:
            disabled: false
            upstream:
              host:
                hostname: newapi.default.svc.cluster.local
                port: 8080
    --- skipped ---
        put:
          summary: Update pet
          description: Update an existing pet by Id
          operationId: updatePet
     --- skipped ---

```
