# The Kusk OpenAPI Extension

Kusk Gateway comes with an `x-kusk` [OpenAPI extension](https://swagger.io/specification/#specification-extensions) to allow
an OpenAPI definition to be the source of truth for both operational and functional aspects of your APIs.

The [extension reference](../reference/extension.md) describes all available properties and the following guides are 
available to help you make the most of them:

- [Mocking](mocking.md) - How to mock all or parts of your API.
- [Validation](validation.md) - How work with automatic request validation.
- [CORS](cors.md) - How to specify CORS settings.
- [Routing](routing.md) - How to configure routing of API requests.
- [Timeouts](timeouts.md) - How to set request timeouts.

## **Properties Overview**

`x-kusk` extension can be applied at (not exclusively):

1. Top level of an OpenAPI definition:
```yaml
  openapi: 3.0.2
  info:
    title: Swagger Petstore - OpenAPI 3.0
  x-kusk:
    hosts:
    - "example.org"
    disabled: false
    cors:
      ...
```

2. Path level:
```yaml
openapi: 3.0.2
info:
  title: Swagger Petstore - OpenAPI 3.0
paths:
  /pet:
    x-kusk:
      disabled: true # disables all /pet endpoints
    post:
      ...
```

3. Method (operation) level:
```yaml
  openapi: 3.0.2
  info:
    title: Swagger Petstore - OpenAPI 3.0
  paths:
    /pet:
      post:
        x-kusk:
          upstream: # routes the POST /pet endpoint to a Kubernetes service
            service:
              namespace: default
              name: petstore
              port: 8000
        ...
```

## **Property Overriding/Inheritance**

The `x-kusk` extension at the operation level takes precedence, or overrides, what is specified at the path level, including the `disabled` option.
Likewise, the path level settings override what is specified at the global level.

If settings aren't specified at a path or operation level, they will be inherited from the layer above, (Operation > Path > Global).
