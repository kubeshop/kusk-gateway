# Authentication

Currently the methods of authentication you can use are any methods that are Header based. That includes Basic authentication, JWT tokens, Bearer tokens, etc.

To use the authentication, you will need to add the following section to your `x-kusk`:

```yaml
openapi: 3.0.0
info:
  title: simple-api
  version: 0.1.0
x-kusk:
  auth:
    type: basic
    auth-upstream:
      host:
        hostname: basic-auth-svc.default
        port:8080
```

The example above authenticates requests to the whole API.

You can also specify different authentication settings for a specific operation or path. The following example shows authentication configuration for a specific operation:

```yaml
...
paths:
  /hello:
    get:
      operationId: getHello
      auth:
        type: basic
        auth-upstream:
          host:
            hostname: basic-auth-svc.default
            port:8080
      ..
```

As you can see in the example, in the `hostname` section we're referencing a service in our cluster that will handle the authentication of the requests. This flow looks like the following: 

![kusk-gateway basic-auth diagram](../img/basic-auth.png)

!!! note non-important "`hostname` can contain any valid domain"

    The `hostname` field can contain internal cluster domains but also external domains to your cluster that can handle your header based authentication.

See all available Authentication configuration options in the [Extension Reference](../../reference/extension/#authentication).
