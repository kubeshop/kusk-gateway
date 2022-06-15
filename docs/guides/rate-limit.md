# Rate limiting

Rate limiting ensures that your application doesn't get more than a specified number of requests over time. It effectively helps to protect your API from overloading. For requests above the threshold, Kusk Gateway returns HTTP Too Many Requests error.

Kusk makes it easy to configure Rate Limiting, using the `rate_limit` option in the `x-kusk` extension:

```yaml
openapi: 3.0.0
info:
  title: simple-api
  version: 0.1.0
x-kusk:
  rate_limit:
    requests_per_unit: 2
    unit: minute
..
```

The example above allows only up to two requests per minute to be sent to the whole API. 

You can also specify different rate-limiting settings for a specific operation or path. The following example shows rate limiting configuration for a specific operation:

```yaml
...
paths:
  /hello:
    get:
      operationId: getHello
      x-kusk:
        rate_limit:
          requests_per_unit: 2
          unit: minute
      ..
```

See all available Rate Limiting configuration options in the [Extension Reference](../../reference/extension/#rate-limiting).
