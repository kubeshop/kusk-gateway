# Caching

The HTTP Cache stores a response associated with a request and reuses the stored response for subsequent requests. Caches reduce latency and network traffic, as the response is directly returned from the gateway. 

Kusk Gateway implements all the complexity of HTTP caching semantics. For more information, read [Envoy's HTTP Caching documentation](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/cache_filter).
 
Kusk makes caching easy to configure using a simple OpenAPI extension:

```yaml
openapi: 3.0.0
info:
  title: simple-api
  version: 0.1.0
x-kusk:
  cache:
    enabled: true
    max_age: 60
..
```

The example above caches responses to HTTP GET requests for 60 seconds. 

You can also specify different caching settings for a specific operation or path. The following example shows rate limiting configuration for a specific operation:

```yaml
...
paths:
  /hello:
    get:
      operationId: getHello
      x-kusk:
        cache:
          enabled: true
          max_age: 60
      ..
```

See all available Caching configuration options in the [Extension Reference](../../reference/extension/#rate-limiting).
